# Web Hunt

This repository contains a port of the real-time multiplayer 2D maze war BSD game [__Hunt__](https://man.openbsd.org/hunt.6)
to Google App Engine and modern[^1] web browsers.  

_The object of the game hunt is to kill off the other players. There are no rooms, no treasures, and no monsters. Instead, you wander around a maze, find grenades, trip mines, and shoot down walls and players. The more players you kill before you die, the better your score is._

Hunt is a moderately complex multiplayer real-time game.  Multiplayer web and mobile games require
a mix of stateless and stateful components to deal with managing game
state, load balancing, scaling, matchmaking, statistics gathering,
monitoring, analytics, and more.

This project implements one such design and architecture for modernizing
the retro terminal based real-time multiplayer 1980's BSD game, "Hunt"
and running in Google Cloud's App Engine serverless platform.

It includes the entire client and server side applications:

* The Single Page App running in your browser
* The stateless serving infrastructure running in the App Engine Standard Environment
* The stateful game engine running in the App Engine Flexible environment.
* Memcache, Datastore, Cloud Pub-Sub, Cron, and Metadata services are utilized as architectural components.

Slides containing more information about this project, including uncovering and fixing a 36 year old "0-Day" bug
are at [Webifying a Retro Realtime Multiplayer Game](https://docs.google.com/presentation/d/1POgLr-T5vJAh_vDTv9yzXvAeQnoigpQPYpYbli7jpq0/edit?usp=sharing)

## Highlights
* Client: HTML / Javascript
* Frontend: Golang application for App Engine Standard
* Backend: Game engine (Original Hunt daemon & Golang instance manager)
  in App Engine Flex

## Licenses
* All Google authored code is covered by the license detailed in the LICENSE
  file at the top of the tree.
* This repository also contains code covered by other licenses.
  All such code can be found within the "third_party" directory tree.

## Installation
### Prerequisites

* Install The Go Programming Language  
	Installation instructions can be found at [https://golang.org](https://golang.org "https://golang.org")

* Install docker  
	Installation instructions can be found at [https://docker.github.io/engine/installation](https://docker.github.io/engine/installation)

* Install the App Engine SDK for Go  
	Installation instructions can be found at [https://cloud.google.com/appengine/docs/go/download](https://cloud.google.com/appengine/docs/go/download)

* Install the Google Cloud SDK  
	Installation instructions can be found at [https://cloud.google.com/sdk/docs/quickstarts](https://cloud.google.com/sdk/docs/quickstarts)
 
### Build and Deploy
* Using the Google Cloud Platform dashboard, go to the Projects page and
create a new project.  
[https://console.cloud.google.com/iam-admin/projects](https://console.cloud.google.com/iam-admin/projects)  
Make note of the name assigned to the new project.

* From the root of the web-bsd-hunt directory, build and deploy the project by executing the command:

     `make PROJECT=<project-name> build deploy`  
     
     Where `<project-name>` is the name given to the Google Platform project created in the previous step.

* Once the build and deployment finishes, you can view the application in a web browser using the following URL:  
 https://`<project-name>`.appspot.com

## History
* I created and open sourced this project as an onboarding exercise to learn the platform when I joined the App Engine team in 2016.
  It was officially open sourced through the Google process, and this fork is intended to be the modernized version now that I no
  longer work at Google.

[^1]: This project was last updated in 2016, so no doubt some work is required to get it working again.  Pull requests happily accepted!
