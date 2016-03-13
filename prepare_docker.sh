#!/bin/bash

cp ./gonzales ./docker

(
	cd docker || exit
	docker build -t gonzales .
)
