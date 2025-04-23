package main

func doIt() {
	// start by initializing a given/sent directory
	// go to the dir and check if it's already a minigit project
	// if yes, failed. if not continues
	// depending on the action do a few things
	// normally, the flow are add,commit, push(???)
	// add
	// walk through the directory, and calc hash from the content of each file
	// this is done to create the name of the hash object
	// and if it's done before, search for existing duplicate on the project by comparing hashes
	// if duplicate found, that mean that file hasn't changed from the previous state and skipped
	// all of the new file are added(staged) if they are different from previous version
	// what actually happen is that the file are tracked, so that when we execute commit
}
