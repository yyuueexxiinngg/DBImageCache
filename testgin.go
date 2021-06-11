package main

import (
	"log"

	"github.com/anacrolix/torrent"
)

func main() {
	c, _ := torrent.NewClient(nil)
	defer c.Close()
	t, _ := c.AddMagnet("magnet:?xt=urn:btih:E9BC30AC1B20014FF8D09665B1F6424C2EDAAB9B")
	<-t.GotInfo()
	t.DownloadAll()
	c.WaitAll()
	log.Print("ermahgerd, torrent downloaded")
}
