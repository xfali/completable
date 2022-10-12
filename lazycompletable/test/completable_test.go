/*
 * Copyright 2022 Xiongfa Li.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package test

import (
	"fmt"
	"github.com/xfali/completable"
	"github.com/xfali/completable/lazycompletable"
	"testing"
	"time"
)

func TestCompletedFuture(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cf := lazycompletable.CompletedFuture("Hello world").ThenAccept(func(s string) {
			if s != "Hello world" {
				t.Fatal("not match")
			}
		})
		cf.Get(nil)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
	})

	t.Run("nil", func(t *testing.T) {
		cf := lazycompletable.CompletedFuture(nil).ThenAccept(func(s string) {
			if s != "" {
				t.Fatal("not match")
			}
		})
		cf.Get(nil)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
	})

	t.Run("nil 2", func(t *testing.T) {
		cf := lazycompletable.CompletedFuture(nil).ThenAccept(func(s *bool) {
			if s != nil {
				t.Fatal("not match")
			}
		})
		cf.Get(nil)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
	})
}

func TestAllOf(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.AllOf(lazycompletable.SupplyAsync(func() int {
			time.Sleep(2 * time.Second)
			return 1
		}), lazycompletable.SupplyAsync(func() int {
			time.Sleep(1 * time.Second)
			return 1
		}), lazycompletable.SupplyAsync(func() int {
			time.Sleep(3 * time.Second)
			return 1
		}))
		cf.Get(nil)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
		useTime := time.Since(now)
		if useTime < 1*time.Second || useTime > 4*time.Second {
			t.Fatal("must 3 second")
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Log("Panic!", r)
				if r.(string) != "error on 2" {
					t.Fatal("not match")
				}
			}
		}()
		now := time.Now()
		cf := lazycompletable.AllOf(lazycompletable.SupplyAsync(func() int {
			time.Sleep(2 * time.Second)
			panic("error on 2")
			return 1
		}), lazycompletable.SupplyAsync(func() int {
			time.Sleep(1 * time.Second)
			return 1
		}), lazycompletable.SupplyAsync(func() int {
			time.Sleep(3 * time.Second)
			return 1
		}))
		cf.Get(nil)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
		useTime := time.Since(now)
		if useTime < 1*time.Second || useTime > 4*time.Second {
			t.Fatal("must 3 second")
		}
	})

	//t.Run("normal cancel", func(t *testing.T) {
	//	now := time.Now()
	//	cf := completablefuture.AllOf(completablefuture.SupplyAsync(func() int {
	//		time.Sleep(2 * time.Second)
	//		return 1
	//	}), completablefuture.SupplyAsync(func() int {
	//		time.Sleep(1 * time.Second)
	//		return 1
	//	}), completablefuture.SupplyAsync(func() int {
	//		time.Sleep(3 * time.Second)
	//		return 1
	//	}))
	//	go func() {
	//		time.Sleep(2 * time.Second)
	//		cf.Cancel()
	//	}()
	//	cf.Get(nil)
	//	if !cf.IsDone() {
	//		t.Fatal("Must be done")
	//	}
	//	useTime := time.Since(now)
	//	if useTime < 1*time.Second || useTime > 2100*time.Millisecond {
	//		t.Fatal("must between 1 - 2 second")
	//	}
	//})
}

func TestAnyOf(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.AnyOf(lazycompletable.SupplyAsync(func() int {
			time.Sleep(2 * time.Second)
			return 2
		}), lazycompletable.SupplyAsync(func() int {
			time.Sleep(1 * time.Second)
			return 1
		}), lazycompletable.SupplyAsync(func() int {
			time.Sleep(3 * time.Second)
			return 3
		}))
		var ret int
		err := cf.Get(&ret)
		if err != nil {
			t.Fatal("not error")
		}
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
		useTime := time.Since(now)
		if useTime < 1*time.Second || useTime > 1100*time.Millisecond {
			t.Fatal("must  1 second")
		}

		if ret != 1 {
			t.Fatal("must be 1")
		}
	})

	t.Run("panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Log("Panic!", r)
				if r.(string) != "1 second panic" {
					t.Fatal("not match")
				}
			}
		}()
		now := time.Now()
		cf := lazycompletable.AnyOf(lazycompletable.SupplyAsync(func() int {
			time.Sleep(2 * time.Second)
			panic("2 second panic")
			return 2
		}), lazycompletable.SupplyAsync(func() int {
			time.Sleep(1 * time.Second)
			panic("1 second panic")
			return 1
		}), lazycompletable.SupplyAsync(func() int {
			time.Sleep(3 * time.Second)
			return 3
		}))
		cf.Get(nil)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
		useTime := time.Since(now)
		if useTime < 1*time.Second || useTime > 1100*time.Millisecond {
			t.Fatal("must  1 second")
		}
	})

	//t.Run("cancel", func(t *testing.T) {
	//	now := time.Now()
	//	cf := completablefuture.AnyOf(completablefuture.SupplyAsync(func() int {
	//		time.Sleep(2 * time.Second)
	//		return 1
	//	}), completablefuture.SupplyAsync(func() int {
	//		time.Sleep(1 * time.Second)
	//		return 1
	//	}), completablefuture.SupplyAsync(func() int {
	//		time.Sleep(3 * time.Second)
	//		return 1
	//	}))
	//	go func() {
	//		time.Sleep(200 * time.Millisecond)
	//		cf.Cancel()
	//	}()
	//	cf.Get(nil)
	//	if !cf.IsDone() {
	//		t.Fatal("Must be done")
	//	}
	//	useTime := time.Since(now)
	//	if useTime > 290*time.Millisecond {
	//		t.Fatal("must  200 mill")
	//	}
	//})
}

func TestRunAsync(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.RunAsync(func() {
			time.Sleep(1 * time.Second)
			t.Log(time.Now().Sub(now), "Hello World")
		})
		cf.Get(nil)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
	})

	t.Run("with cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.RunAsync(func() {
			time.Sleep(1 * time.Second)
			t.Log(time.Now().Sub(now), "Hello World")
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		cf.Get(nil)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
}

func TestSupplyAsync(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		})
		ret := ""
		err := cf.Get(&ret)
		t.Log("after get1: ", time.Now().Sub(now))
		if err != nil {
			t.Fatal(err)
		}
		// get twice
		err = cf.Get(&ret)
		t.Log("after get2: ", time.Now().Sub(now))
		if err != nil {
			t.Fatal(err)
		}
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
		t.Log(time.Now().Sub(now), ret)
		if ret != "Hello world" {
			t.Fatal("not match")
		}
	})

	t.Run("with cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		ret := ""
		err := cf.Get(&ret)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(err, ret)
		t.Log(time.Now().Sub(now), ret)
		if ret == "Hello world" {
			t.Fatal("not match")
		}
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
}

func TestThenApply(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenApply(func(o string) string {
			return o + " world"
		})
		ret := ""
		cf.Get(&ret)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
		t.Log(time.Now().Sub(now), ret)
		if ret != "Hello world" {
			t.Fatal("not match")
		}
	})
	t.Run("sync with cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.ThenApply(func(o string) string {
			return o + " world"
		})
		ret := ""
		err := cf.Get(&ret)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if ret == "Hello world" {
			t.Fatal("not match")
		}
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenApplyAsync(func(o string) string {
			time.Sleep(1 * time.Second)
			return o + " world"
		})
		cf.Get(&ret)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
		t.Log(time.Now().Sub(now), ret)
		if ret != "Hello world" {
			t.Fatal("not match")
		}
	})

	t.Run("async cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenApplyAsync(func(o string) string {
			time.Sleep(1 * time.Second)
			return o + " world"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()

		ret := ""
		err := cf.Get(&ret)
		if !cf.IsDone() {
			t.Fatal("Must be done")
		}
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if ret == "Hello world" {
			t.Fatal("not match")
		}
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
}

func TestThenAccept(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		}).ThenAccept(func(o string) {
			if o != "Hello world" {
				t.Fatal("not match")
			} else {
				t.Log(o)
			}
		})
		t.Log("before get: ", time.Now().Sub(now))
		cf.Get(nil)
		t.Log("after get: ", time.Now().Sub(now))
	})

	t.Run("sync cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.ThenAccept(func(o string) {
			if o != "Hello world" {
				t.Fatal("not match")
			} else {
				t.Log(o)
			}
		})
		ret := ""
		t.Log("before get: ", time.Now().Sub(now))
		err := cf.Get(&ret)
		t.Log("after get: ", time.Now().Sub(now))
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})

	t.Run("async", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		}).ThenAcceptAsync(func(o string) {
			if o != "Hello world" {
				t.Fatal("not match")
			} else {
				t.Log(o)
			}
		})
		t.Log("before get: ", time.Now().Sub(now))
		cf.Get(nil)
		t.Log("after get: ", time.Now().Sub(now))
		t.Log(time.Now().Sub(now))
	})

	t.Run("async cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		}).ThenAcceptAsync(func(o string) {
			if o != "Hello world" {
				t.Fatal("not match")
			} else {
				t.Log(o)
			}
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		ret := ""
		err := cf.Get(&ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(err, ret)
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
}

func TestThenRun(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		}).ThenRun(func() {
			t.Log("in", time.Now().Sub(now))
		})
		cf.Get(nil)
		t.Log(time.Now().Sub(now))
	})
	t.Run("sync cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.ThenRun(func() {
			t.Log("in", time.Now().Sub(now))
		})
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		}).ThenRunAsync(func() {
			t.Log("in", time.Now().Sub(now))
		})
		cf.Get(nil)
		t.Log(time.Now().Sub(now))
	})

	t.Run("async cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		}).ThenRunAsync(func() {
			t.Log("in", time.Now().Sub(now))
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
}

func TestThenCombine(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenCombine(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s1, s2 string) string {
			if time.Since(now) < 2*time.Second {
				t.Fatal("not match")
			}
			return s1 + s2
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != "Hello world" {
			t.Fatal("not match")
		}
	})

	t.Run("sync cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.ThenCombine(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s1, s2 string) string {
			if time.Since(now) < 2*time.Second {
				t.Fatal("not match")
			}
			return s1 + s2
		})
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if ret == "Hello world" {
			t.Fatal("not match")
		}
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})

	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenCombineAsync(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s1, s2 string) string {
			if time.Since(now) < 2*time.Second {
				t.Fatal("not match")
			}
			return s1 + s2
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != "Hello world" {
			t.Fatal("not match")
		}
	})

	t.Run("async cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenCombineAsync(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s1, s2 string) string {
			if time.Since(now) < 2*time.Second {
				t.Fatal("not match")
			}
			return s1 + s2
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if ret == "Hello world" {
			t.Fatal("not match")
		}
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
}

func TestAcceptBoth(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenAcceptBoth(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s1, s2 string) {
			if s1+s2 != "Hello world" {
				t.Fatal("not match")
			}
			if time.Since(now) < 2*time.Second {
				t.Fatal("not match")
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})

	t.Run("sync cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.ThenAcceptBoth(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s1, s2 string) {
			if s1+s2 != "Hello world" {
				t.Fatal("not match")
			}
			if time.Since(now) < 2*time.Second {
				t.Fatal("not match")
			}
		})
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})

	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenAcceptBothAsync(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s1, s2 string) {
			if s1+s2 != "Hello world" {
				t.Fatal("not match")
			}
			if time.Since(now) < 2*time.Second {
				t.Fatal("not match")
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})

	t.Run("async cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenAcceptBothAsync(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s1, s2 string) {
			if s1+s2 != "Hello world" {
				t.Fatal("not match")
			}
			if time.Since(now) < 2*time.Second {
				t.Fatal("not match")
			}
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
}

func TestRunAfterBoth(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).RunAfterBoth(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func() {
			if time.Since(now) < 2*time.Second {
				t.Fatal("not match")
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})

	t.Run("sync cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.RunAfterBoth(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func() {
			if time.Since(now) < 2*time.Second {
				t.Fatal("not match")
			}
		})
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})

	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).RunAfterBothAsync(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func() {
			if time.Since(now) < 2*time.Second {
				t.Fatal("not match")
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})

	t.Run("async cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).RunAfterBothAsync(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func() {
			if time.Since(now) < 2*time.Second {
				t.Fatal("not match")
			}
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
}

func TestApplyEither(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ApplyToEither(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s string) string {
			return s
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != "Hello" {
			t.Fatal("not match")
		}
	})

	t.Run("sync cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.ApplyToEither(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s string) string {
			return s
		})
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
		if ret == "Hello" {
			t.Fatal("not match")
		}
	})

	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ApplyToEitherAsync(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s string) string {
			return s
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != "Hello" {
			t.Fatal("not match")
		}
	})

	t.Run("async cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ApplyToEitherAsync(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s string) string {
			return s
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
		if ret == "Hello" {
			t.Fatal("not match")
		}
	})
}

func TestAcceptEither(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).AcceptEither(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s string) {
			if s != "Hello" {
				t.Fatal("not match")
			} else {
				t.Log(s)
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})

	t.Run("sync cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.AcceptEither(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s string) {
			if s != "Hello" {
				t.Fatal("not match")
			} else {
				t.Log(s)
			}
		})
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})

	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).AcceptEitherAsync(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s string) {
			if s != "Hello" {
				t.Fatal("not match")
			} else {
				t.Log(s)
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})

	t.Run("async cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).AcceptEitherAsync(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func(s string) {
			if s != "Hello" {
				t.Fatal("not match")
			} else {
				t.Log(s)
			}
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
}

func TestRunEither(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).RunAfterEither(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func() {
			x := time.Since(now)
			if time.Second <= x && x < 2*time.Second {
				t.Log(x)
			} else {
				t.Fatal("not match")
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)

	})

	t.Run("sync cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.RunAfterEither(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func() {
			x := time.Since(now)
			if time.Second <= x && x < 2*time.Second {
				t.Log(x)
			} else {
				t.Fatal("not match")
			}
		})
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})

	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).RunAfterEitherAsync(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func() {
			x := time.Since(now)
			if time.Second <= x && x < 2*time.Second {
				t.Log(x)
			} else {
				t.Fatal("not match")
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})

	t.Run("async cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).RunAfterEitherAsync(lazycompletable.SupplyAsync(func() string {
			time.Sleep(2 * time.Second)
			return " world"
		}), func() {
			x := time.Since(now)
			if time.Second <= x && x < 2*time.Second {
				t.Log(x)
			} else {
				t.Fatal("not match")
			}
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
}

func TestCompose(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenCompose(func(s string) completable.CompletionStage {
			t.Log(s)
			if s != "Hello" {
				t.Fatal("not match")
			}
			return lazycompletable.SupplyAsync(func() string {
				return "world"
			})
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != "world" {
			t.Fatal("not match")
		}
	})

	t.Run("sync cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.ThenCompose(func(s string) completable.CompletionStage {
			t.Log(s)
			if s != "Hello" {
				t.Fatal("not match")
			}
			return lazycompletable.SupplyAsync(func() string {
				return "world"
			})
		})
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
		if ret == "world" {
			t.Fatal("not match")
		}
	})

	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenComposeAsync(func(s string) completable.CompletionStage {
			t.Log(s)
			if s != "Hello" {
				t.Fatal("not match")
			}
			return lazycompletable.SupplyAsync(func() string {
				return "world"
			})
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != "world" {
			t.Fatal("not match")
		}
	})

	t.Run("async cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenComposeAsync(func(s string) completable.CompletionStage {
			t.Log(s)
			if s != "Hello" {
				t.Fatal("not match")
			}
			return lazycompletable.SupplyAsync(func() string {
				return "world"
			})
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
		if ret == "world" {
			t.Fatal("not match")
		}
	})
}

func TestExceptionally(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error!")
			return "Hello"
		}).Exceptionally(func(o interface{}) string {
			t.Log(o)
			return "world"
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != "world" {
			t.Fatal("not match")
		}
	})

	t.Run("sync cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error!")
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.Exceptionally(func(o interface{}) string {
			t.Log(o)
			return "world"
		})
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
		if ret == "world" {
			t.Fatal("not match")
		}
	})

	t.Run("sync no panic", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).Exceptionally(func(o interface{}) string {
			t.Log(o)
			return "world"
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != "Hello" {
			t.Fatal("not match")
		}
	})

	t.Run("sync no panic cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.Exceptionally(func(o interface{}) string {
			t.Log(o)
			return "world"
		})
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
		if ret == "Hello" {
			t.Fatal("not match")
		}
	})
}

func TestWhenComplete(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error")
			return "Hello"
		}).WhenComplete(func(s string, panic interface{}) {
			t.Log(panic)
			if s != "" || panic.(string) != "error" {
				t.Fatal("not match")
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})

	t.Run("sync cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error")
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.WhenComplete(func(s string, panic interface{}) {
			t.Log(panic)
			//have been cancelled
			//if s != "" || panic.(string) != "error" {
			//	t.Fatal("not match")
			//}
		})
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})

	t.Run("sync no panic", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).WhenComplete(func(s string, panic interface{}) {
			t.Log(s, panic)
			if s != "Hello" || panic != nil {
				t.Fatal("not match")
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})

	t.Run("sync no panic cancel", func(t *testing.T) {
		now := time.Now()
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.WhenComplete(func(s string, panic interface{}) {
			t.Log(s, panic)
			//if s != "Hello" || panic != nil {
			//	t.Fatal("not match")
			//}
		})

		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})

	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error")
			return "Hello"
		}).WhenCompleteAsync(func(s string, panic interface{}) {
			t.Log(panic)
			if s != "" || panic.(string) != "error" {
				t.Fatal("not match")
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})

	t.Run("async cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error")
			return "Hello"
		}).WhenCompleteAsync(func(s string, panic interface{}) {
			t.Log(s, panic)
			//if s != "" || panic.(string) != "error" {
			//	t.Fatal("not match")
			//}
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})

	t.Run("async no panic", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).WhenCompleteAsync(func(s string, panic interface{}) {
			t.Log(s, panic)
			if s != "Hello" || panic != nil {
				t.Fatal("not match")
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})

	t.Run("async no panic cancel", func(t *testing.T) {
		now := time.Now()
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).WhenCompleteAsync(func(s string, panic interface{}) {
			t.Log(s, panic)
			//if s != "Hello" || panic != nil {
			//	t.Fatal("not match")
			//}
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		ret := ""
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
	})
}

func TestHandle(t *testing.T) {
	t.Run("sync panic", func(t *testing.T) {
		now := time.Now()
		ret := 0
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error")
			return "Hello"
		}).Handle(func(s string, panic interface{}) int {
			t.Log(panic)
			if s != "" || panic.(string) != "error" {
				t.Fatal("not match")
			}
			if s != "" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != 2 {
			t.Fatal("not match")
		}
	})

	t.Run("sync panic cancel", func(t *testing.T) {
		now := time.Now()
		ret := 0
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error")
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.Handle(func(s string, panic interface{}) int {
			t.Log(panic)
			//if s != "" || panic.(string) != "error" {
			//	t.Fatal("not match")
			//}
			if s != "" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
		if ret == 2 {
			t.Fatal("not match")
		}
	})

	t.Run("sync no panic", func(t *testing.T) {
		now := time.Now()
		ret := 0
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).Handle(func(s string, panic interface{}) int {
			t.Log(panic)
			if s != "Hello" || panic != nil {
				t.Fatal("not match")
			}
			if s != "" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != 1 {
			t.Fatal("not match")
		}
	})

	t.Run("sync no panic cancel", func(t *testing.T) {
		now := time.Now()
		ret := 0
		origin := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			origin.Cancel()
		}()
		cf := origin.Handle(func(s string, panic interface{}) int {
			t.Log(panic)
			//if s != "Hello" || panic != nil {
			//	t.Fatal("not match")
			//}
			if s != "" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
		if ret == 1 {
			t.Fatal("not match")
		}
	})

	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := 0
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error")
			return "Hello"
		}).HandleAsync(func(s string, panic interface{}) int {
			t.Log(panic)
			if s != "" || panic.(string) != "error" {
				t.Fatal("not match")
			}
			if s != "" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != 2 {
			t.Fatal("not match")
		}
	})

	t.Run("async cancel", func(t *testing.T) {
		now := time.Now()
		ret := 0
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error")
			return "Hello"
		}).HandleAsync(func(s string, panic interface{}) int {
			t.Log(s, panic)
			//if s != "" || panic.(string) != "error" {
			//	t.Fatal("not match")
			//}
			if s != "" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
		if ret == 2 {
			t.Fatal("not match")
		}
	})

	t.Run("async no panic", func(t *testing.T) {
		now := time.Now()
		ret := 0
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).HandleAsync(func(s string, panic interface{}) int {
			t.Log(panic)
			if s != "Hello" || panic != nil {
				t.Fatal("not match")
			}
			if s != "" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != 1 {
			t.Fatal("not match")
		}
	})

	t.Run("async no panic cancel", func(t *testing.T) {
		now := time.Now()
		ret := 0
		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).HandleAsync(func(s string, panic interface{}) int {
			t.Log(s, panic)
			//if s != "" || panic.(string) != "error" {
			//	t.Fatal("not match")
			//}
			if s != "" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Cancel()
		}()
		err := cf.Get(&ret)
		t.Log(err, ret)
		if !cf.IsCancelled() {
			t.Fatal("must be cancelled")
		}
		t.Log(time.Now().Sub(now), ret)
		if time.Since(now) >= 1*time.Second {
			t.Fatal("have be cancelled less 1 second")
		}
		if ret == 1 {
			t.Fatal("not match")
		}
	})
}

func TestComplete(t *testing.T) {
	t.Run("sync panic", func(t *testing.T) {
		now := time.Now()
		ret := 0

		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error")
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Complete("complete")
		}()
		cf = cf.Handle(func(s string, panic interface{}) int {
			t.Log(panic)
			if s != "complete" || panic != nil {
				t.Fatal("not match")
			}
			if s == "complete" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != 1 {
			t.Fatal("not match")
		}
	})

	t.Run("sync no panic", func(t *testing.T) {
		now := time.Now()
		ret := 0

		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.Complete("complete")
		}()
		cf = cf.Handle(func(s string, panic interface{}) int {
			t.Log(panic)
			if s != "complete" || panic != nil {
				t.Fatal("not match")
			}
			if s == "complete" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != 1 {
			t.Fatal("not match")
		}
	})

	t.Run("async panic", func(t *testing.T) {
		now := time.Now()
		ret := 0

		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			fmt.Println("done1")
			panic("error")
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			fmt.Println("done2")
			cf.Complete("complete")
		}()
		cf2 := cf.HandleAsync(func(s string, panic interface{}) int {
			t.Log(panic)
			if s != "complete" || panic != nil {
				t.Fatal("not match")
			}
			if s == "complete" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		cf2.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != 1 {
			t.Fatal("not match")
		}
	})

	t.Run("async no panic", func(t *testing.T) {
		now := time.Now()
		ret := 0

		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			fmt.Println("done1")
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			fmt.Println("done2")
			cf.Complete("complete")
		}()
		cf2 := cf.HandleAsync(func(s string, panic interface{}) int {
			t.Log(panic)
			if s != "complete" || panic != nil {
				t.Fatal("not match")
			}
			if s == "complete" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		cf2.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != 1 {
			t.Fatal("not match")
		}
	})
}

func TestCompleteExceptionally(t *testing.T) {
	t.Run("sync panic", func(t *testing.T) {
		now := time.Now()
		ret := 0

		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error")
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.CompleteExceptionally("complete")
		}()
		cf = cf.Handle(func(s string, panic interface{}) int {
			t.Log(panic)
			if s != "" || panic.(string) != "complete" {
				t.Fatal("not match")
			}
			if s == "Hello" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != 2 {
			t.Fatal("not match")
		}
	})

	t.Run("sync no panic", func(t *testing.T) {
		now := time.Now()
		ret := 0

		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.CompleteExceptionally("complete")
		}()
		cf = cf.Handle(func(s string, panic interface{}) int {
			t.Log(panic)
			if s != "" || panic.(string) != "complete" {
				t.Fatal("not match")
			}
			if s == "Hello" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != 2 {
			t.Fatal("not match")
		}
	})

	t.Run("async panic", func(t *testing.T) {
		now := time.Now()
		ret := 0

		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			panic("error")
			return "Hello"
		})
		cf2 := cf.HandleAsync(func(s string, panic interface{}) int {
			t.Log(panic)
			if s != "" || panic.(string) != "complete" {
				t.Fatal("not match")
			}
			if s == "Hello" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf2.CompleteExceptionally("complete")
		}()
		cf2.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != 2 {
			t.Fatal("not match")
		}
	})

	t.Run("async panic", func(t *testing.T) {
		now := time.Now()
		ret := 0

		cf := lazycompletable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			cf.CompleteExceptionally("complete")
		}()
		cf2 := cf.HandleAsync(func(s string, panic interface{}) int {
			t.Log(panic)
			if s != "" || panic.(string) != "complete" {
				t.Fatal("not match")
			}
			if s == "Hello" {
				return 1
			}
			if panic != nil {
				return 2
			}
			return 0
		})
		cf2.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != 2 {
			t.Fatal("not match")
		}
	})
}
