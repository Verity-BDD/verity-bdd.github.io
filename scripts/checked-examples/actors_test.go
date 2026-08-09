package checkedexamples

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	verity "github.com/verity-bdd/verity-bdd"
)

type actorAbility struct{ owner string }

func TestActorsLifecycleExample(t *testing.T) {
	var factoryCalls atomic.Int32
	test := verity.NewVerityTest(t, verity.Scene{
		DefaultAbilities: []verity.DefaultAbilityFactory{
			func(name string) verity.Ability {
				factoryCalls.Add(1)
				return &actorAbility{owner: name}
			},
		},
	})

	var wg sync.WaitGroup
	for _, name := range []string{"alice", "Bob", "Alice", "Bob"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			test.ActorCalled(name)
		}()
	}
	wg.Wait()

	bob := test.ActorCalled("Bob")
	if bob != test.ActorCalled("Bob") {
		t.Fatal("same name must return the original actor")
	}
	actors := test.Actors()
	if actors == nil {
		t.Fatal("Actors must return a non-nil snapshot")
	}
	got := []string{actors[0].Name(), actors[1].Name(), actors[2].Name()}
	if !reflect.DeepEqual(got, []string{"Alice", "Bob", "alice"}) {
		t.Fatalf("unexpected case-sensitive order: %v", got)
	}
	if actors[1] != bob || factoryCalls.Load() != 3 {
		t.Fatal("snapshot identity or default factory reuse contract violated")
	}
	actors[0] = nil
	if test.Actors()[0] == nil {
		t.Fatal("mutating a snapshot changed the registry")
	}

	test.Shutdown()
	if actors := test.Actors(); actors == nil || len(actors) != 0 {
		t.Fatalf("Actors after shutdown = %#v, want non-nil empty snapshot", actors)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("ActorCalled after Shutdown must panic")
		}
	}()
	test.ActorCalled("Too late")
}
