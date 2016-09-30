// Copyright 2016 The Web BSD Hunt Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
////////////////////////////////////////////////////////////////////////////////
//
// TODO: High-level file comment.
package main

import(
	"fmt"
	"net/http"
	"time"

	"apputils"
	"gamerpc"

	"google.golang.org/appengine"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/datastore"
)

const(
	GameInstanceTimeout	= 2 * time.Minute
)

// TODO(tad): this is here because Datastore can't handle the types in a gamerpc.GameClient.  Ugh.
type DatastoreGameInfo struct {
	URL	string
}

type GameInstance struct {
	InstanceID	string
	URL		string
}

func FindGameInstance(r *http.Request, urlstr string) (*gamerpc.GameClient, error) {
	if staticGameClient != nil {
		return staticGameClient, nil
	}

	ctx := appengine.NewContext(r)
	game := gamerpc.GameClient{}
	_, err := memcache.JSON.Get(ctx, urlstr, &game)
	if err != nil {
		return nil, err
	}

	return &game, nil
}

func UpdateGameInstance(r *http.Request, instance string, urlstr string) (*gamerpc.GameClient, error) {
	ctx := appengine.NewContext(r)

	game, err := gamerpc.NewGameClient(urlstr, rpcTypeStr, rpOptions)
	if err != nil {
		return nil, err
	}

	item := &memcache.Item {
		Key:		urlstr,
		Object:		game,
		Expiration:     GameInstanceTimeout,
	}

	// NOTE: we don't bother with CAS because it doesn't really matter who wins
	err = memcache.JSON.Set(ctx, item)
	if err != nil {
		return nil, err
	}

	apputils.Log(r, fmt.Sprintf("UpdateGameInstance[%s]: Memcache updated with %s: %v", instance, urlstr, item))

	dgame := &DatastoreGameInfo {
		URL: urlstr,
	}

	instanceKey := datastore.NewKey(ctx, "instances", instance, 0, nil)
	key, err := datastore.Put(ctx, instanceKey, dgame)
	if err != nil {
		return nil, err
	}

	apputils.Log(r, fmt.Sprintf("UpdateGameInstance[%s]: Datastore updated with key %s gameserver %s", instance, key, dgame.URL))

	return game, nil
}

//
// Deletes all instances found in the database that aren't in memcache.
// This technique is used because memcache is setup to expire the instances
// if not heard from for too long.
//
func ReapGameInstances(r *http.Request) (int, error) {
	apputils.Log(r, "ReapGameInstances: start")

	ctx := appengine.NewContext(r)
	query := datastore.NewQuery("instances")

	var urls []*DatastoreGameInfo

	keys, err := query.GetAll(ctx, &urls)
	if err != nil {
		apputils.Log(r, fmt.Sprintf("Ignore error finding game instances to reap: %v", err))
		return 0, nil
	}

	n := 0
	for i, dgame := range urls {
		apputils.Log(r, fmt.Sprintf("ReapGameInstances[%d]: Look for key %s url %s", i, keys[i], dgame.URL))
		_, err := FindGameInstance(r, dgame.URL)
		if err == nil {
			apputils.Log(r, fmt.Sprintf("ReapGameInstances[%d]: Keep: key %s url %s still alive", i, keys[i], dgame.URL))
			continue
		}

		err = datastore.Delete(ctx, keys[i])
		if err != nil {
			apputils.Log(r, fmt.Sprintf("ReapGameInstances[%d]: Ignore error deleting key %s url %s: %v", i, keys[i], dgame.URL, err))
		} else {
			apputils.Log(r, fmt.Sprintf("ReapGameInstances[%d]: Deleted key %s game %s", i, keys[i], dgame.URL))
			n++
		}
	}

	return n, nil
}

func GameInstances(r *http.Request) ([]*GameInstance, error) {
	var instances []*GameInstance

	apputils.Log(r, "GameInstances: start")
	if staticGameClient != nil {
		instances = append(instances, &GameInstance{InstanceID: "0", URL:gameURLStr})
	} else {
		ctx := appengine.NewContext(r)
		query := datastore.NewQuery("instances")

		var gameinfo []*DatastoreGameInfo
		keys, err := query.GetAll(ctx, &gameinfo)
		if err != nil {
			return nil, err
		}

		for i, key := range keys {
			id := key.StringID()
			if id == "" {
				apputils.Log(r, fmt.Sprintf("Key %v missing stringid", key))
				continue
			}

			instance := &GameInstance{
				InstanceID:	id,
				URL:		gameinfo[i].URL,
			}
			instances = append(instances, instance)
		}
	}

	apputils.Log(r, fmt.Sprintf("GameInstances: return %d instances", len(instances)))

	return instances, nil
}
