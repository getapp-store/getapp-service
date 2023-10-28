package handlers

import "net/http"

type Deployer struct {
}

func (d *Deployer) Upload(w http.ResponseWriter, r *http.Request) {
	// get upload file from url
	// make apk url
	// upload to rustore
	// upload to another store
}
