#Web Hunt
Port of the real-time multiplayer 2D maze war BSD game "Hunt" to Google
App Engine and modern web browsers.

[https://en.wikipedia.org/wiki/Hunt_(video_game)](https://en.wikipedia.org/wiki/Hunt_(video_game))

Moderately complex multiplayer real-time web and mobile games require
a mix of stateless and stateful components to deal with managing game
state, load balancing, scaling, matchmaking, statistics gathering,
monitoring, analytics, and more.

This project implements one such design and architecture for modernizing
the retro terminal based real time multiplayer 1980's BSD game, "Hunt"
and running in Google App Engine Platform as a Service.

It includes the entire client and server side applications, from the
Javascript Single Page App running in a browser, to the stateless
serving infrastructure running in the Google App Engine (GAE) Standard
Environment, a stateful game engine in the GAE Flexible environment.
Memcache, Datastore, Cloud Pub-Sub, Cron, and Metadata services are
utilized as architectural components.

##Highlights
* Client: HTML / Javascript
* Frontend: Golang application for App Engine Standard
* Backend: Game engine (Original Hunt daemon & Golang instance manager)
  in App Engine Flex

##Licenses
* All Google authored code is covered by the license detailed in the LICENSE
  file at the top of the tree.
* This repository also contains code covered by other licenses.  All
  such code can be found within the "third_party" directory tree.

##Notes
* This is not an official Google product.
