package context

import (
	"sync"
	"testing"
)

func TestContexts(t *testing.T) {
	mgr1 := NewContextManager()
	mgr2 := NewContextManager()

	CheckVal := func(mgr *ContextManager, key, exp_val string) {
		val, ok := mgr.GetValue(key)
		if len(exp_val) == 0 {
			if ok {
				t.Fatalf("expected no value for key %s, got %s", key, val)
			}
			return
		}
		if !ok {
			t.Fatalf("expected value %s for key %s, got no value",
				exp_val, key)
		}
		if exp_val != val {
			t.Fatalf("expected value %s for key %s, got %s", exp_val, key,
				val)
		}

	}

	Check := func(exp_m1v1, exp_m1v2, exp_m2v1, exp_m2v2 string) {
		CheckVal(mgr1, "key1", exp_m1v1)
		CheckVal(mgr1, "key2", exp_m1v2)
		CheckVal(mgr2, "key1", exp_m2v1)
		CheckVal(mgr2, "key2", exp_m2v2)
	}

	Check("", "", "", "")
	mgr2.AddValues(Values{"key1": "val1c"}, func() {
		Check("", "", "val1c", "")
		mgr1.AddValues(Values{"key1": "val1a"}, func() {
			Check("val1a", "", "val1c", "")
			mgr1.AddValues(Values{"key2": "val1b"}, func() {
				Check("val1a", "val1b", "val1c", "")
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					Check("", "", "", "")
				}()
				Go(func() {
					defer wg.Done()
					Check("val1a", "val1b", "val1c", "")
				})
				wg.Wait()
			})
		})
	})
}
