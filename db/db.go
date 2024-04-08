package db

import "github.com/rosedblabs/rosedb/v2"

var DB *rosedb.DB

func init() {
	// specify the options
	options := rosedb.DefaultOptions
	options.DirPath = "/tmp/rosedb_basic"

	// open a database
	var err error
	DB, err = rosedb.Open(options)
	if err != nil {
		panic(err)
	}
	//defer func() {
	//	_ = DB.Close()
	//}()

	//// set a key
	//err = DB.Put([]byte("name"), []byte("rosedb"))
	//if err != nil {
	//	panic(err)
	//}

	//// get a key
	//val, err := DB.Get([]byte("name"))
	//if err != nil {
	//	panic(err)
	//}
	//println(string(val))

	//// delete a key
	//err = DB.Delete([]byte("name"))
	//if err != nil {
	//	panic(err)
	//}
}
