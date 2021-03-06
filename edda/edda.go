// Package edda Logs the architecture configuration (nodes and links) as it evolves
package edda

import (
	"fmt"
	"github.com/adrianco/spigo/archaius"
	"github.com/adrianco/spigo/collect"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/graphjson"
	"github.com/adrianco/spigo/graphml"
	"github.com/adrianco/spigo/names"
	"log"
	"sync"
	"time"
)

// Logchan is a buffered channel for sending logging messages to, or nil if logging is off
// Created before edda starts so that messages can be buffered without depending on edda schedule
var Logchan chan gotocol.Message
var Wg sync.WaitGroup

// Start edda, to listen for logging data from services
func Start(name string) {
	// use a waitgroup so whoever starts edda can tell the logs have been flushed
	Wg.Add(1)
	defer Wg.Done()
	if Logchan == nil {
		return
	}
	var msg gotocol.Message
	microservices := make(map[string]bool, archaius.Conf.Dunbar)
	var ok bool
	hist := collect.NewHist(name)
	log.Println(name + ": starting")
	if archaius.Conf.GraphmlFile != "" {
		graphml.Enabled = true
	}
	if archaius.Conf.GraphjsonFile != "" {
		graphjson.Enabled = true
	}
	ss := ""
	if archaius.Conf.StopStep > 0 {
		ss = fmt.Sprintf("%v", archaius.Conf.StopStep)
	}
	graphml.Setup(archaius.Conf.GraphmlFile + ss)
	graphjson.Setup(archaius.Conf.GraphjsonFile + ss)
	for {
		msg, ok = <-Logchan
		collect.Measure(hist, time.Since(msg.Sent))
		if !ok {
			break // channel was closed
		}
		if archaius.Conf.Msglog {
			log.Printf("%v(backlog %v): %v\n", name, len(Logchan), msg)
		}
		if msg.Imposition == gotocol.Inform {
			graphml.WriteEdge(msg.Intention)
			graphjson.WriteEdge(msg.Intention)
		} else {
			if msg.Imposition == gotocol.Put {
				if microservices[msg.Intention] == false {
					microservices[msg.Intention] = true
					graphml.WriteNode(msg.Intention + " " + names.Package(msg.Intention))
					graphjson.WriteNode(msg.Intention + " " + names.Package(msg.Intention))
				}
			}
		}
	}
	log.Println(name + ": closing")
	graphml.Close()
	graphjson.Close()
}
