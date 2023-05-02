#!/bin/bash

kill `lsof -t -i:8080`
kill `lsof -t -i:8081`
kill `lsof -t -i:8082`
kill `lsof -t -i:5555`