# My First Go Program

Mostly an exercise to see if Go is usable (for me) as a Ruby+Sinatra replacement. Hence, lots of bad go. Idioms? What idioms?

Currenly (as of 07/02/2016) [deployed live](https://hidden-wave-75799.herokuapp.com/) for playing with.

## installing (on Windows)

* download the latest stable golang version from https://golang.org/doc/install
* clone this repo
* ensure you have your $GOPATH set up ok, in particular with this project sitting in $GOPATH/src/github.com/upthebuzzard/alignment
* get and install the 3rd party import (which ends up in $GOPATH/src/... etc)
   * $ go get     github.com/joho/godotenv
   * $ go install github.com/joho/godotenv
* create a .env file and add a valid FT Search API key
   * SAPI_KEY=...

## building and running

* $ go install github.com/upthebuzzard/alignment
* $ $GOPATH/bin/alignment.exe

## deploying to heroku

* first get godep installed (follow [Heroku's instructions](https://devcenter.heroku.com/articles/deploying-go))
* invoke godep 
   * $ godep save -r ./...
* add/commit the new folder+files or changes thereto
* push to heroku

## Notes

* The article data is taken from the Financial Times' Search API. It is the most recent 100 (max) articles which contain the user-supplied phrase.
* Error checking? Nope.
