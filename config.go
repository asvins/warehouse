package main

// Config struct for this service
type Config struct {
	Server struct {
		Addr string
		Port string
	}
	Database struct {
		User    string
		DbName  string
		SSLMode string
	}
}
