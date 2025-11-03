package hub

import "fmt"

type SourceFactory func(configJSON string) (SourceAdapter, error)
type DestinationFactory func(configJSON string) (DestinationAdapter, error)

var sourceFactories = make(map[string]SourceFactory)
var destinationFactories = make(map[string]DestinationFactory)

func RegisterSourceAdapter(typ string, factory SourceFactory) {
	if _, exists := sourceFactories[typ]; exists {
		panic(fmt.Sprintf("Source adapter type '%s' already registered", typ))
	}
	sourceFactories[typ] = factory
}

func RegisterDestinationAdapter(typ string, factory DestinationFactory) {
	if _, exists := destinationFactories[typ]; exists {
		panic(fmt.Sprintf("Destination adapter type '%s' already registered", typ))
	}
	destinationFactories[typ] = factory
}
