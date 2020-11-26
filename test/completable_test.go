// Copyright (C) 2019-2020, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"github.com/xfali/completable"
	"testing"
	"time"
)

func TestRunAsync(t *testing.T) {
	now := time.Now()
	cf := completable.RunAsync(func() {
		time.Sleep(1 * time.Second)
		t.Log(time.Now().Sub(now), "Hello World")
	})
	cf.Get(nil)
}

func TestSupplyAsync(t *testing.T) {
	now := time.Now()
	cf := completable.SupplyAsync(func() string {
		time.Sleep(1 * time.Second)
		return "Hello world"
	})
	ret := ""
	cf.Get(&ret)
	t.Log(time.Now().Sub(now), ret)
	if ret != "Hello world" {
		t.Fatal("not match")
	}
}

func TestThenApply(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenApply(func(o string) string {
			return o + " world"
		})
		ret := ""
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != "Hello world" {
			t.Fatal("not match")
		}
	})
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenApplyAsync(func(o string) string {
			time.Sleep(1 * time.Second)
			return o + " world"
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != "Hello world" {
			t.Fatal("not match")
		}
	})
}

func TestThenAccept(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		}).ThenAccept(func(o string) {
			if o != "Hello world" {
				t.Fatal("not match")
			} else {
				t.Log(o)
			}
		})
		cf.Get(nil)
		t.Log(time.Now().Sub(now))
	})
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		}).ThenAcceptAsync(func(o string) {
			if o != "Hello world" {
				t.Fatal("not match")
			} else {
				t.Log(o)
			}
		})
		cf.Get(nil)
		t.Log(time.Now().Sub(now))
	})
}

func TestThenRun(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		}).ThenRun(func() {
			t.Log("in", time.Now().Sub(now))
		})
		cf.Get(nil)
		t.Log(time.Now().Sub(now))
	})
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello world"
		}).ThenRunAsync(func() {
			t.Log("in", time.Now().Sub(now))
		})
		cf.Get(nil)
		t.Log(time.Now().Sub(now))
	})
}

func TestThenCombine(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenCombine(completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return " world"
		}), func(s1, s2 string) string {
			return s1 + s2
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != "Hello world" {
			t.Fatal("not match")
		}
	})
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenCombineAsync(completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return " world"
		}), func(s1, s2 string) string {
			return s1 + s2
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
		if ret != "Hello world" {
			t.Fatal("not match")
		}
	})
}

func TestAcceptBoth(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenAcceptBoth(completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return " world"
		}), func(s1, s2 string) {
			if s1+s2 != "Hello world" {
				t.Fatal("not match")
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ThenAcceptBothAsync(completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return " world"
		}), func(s1, s2 string) {
			if s1+s2 != "Hello world" {
				t.Fatal("not match")
			}
		})
		cf.Get(&ret)
		t.Log(time.Now().Sub(now), ret)
	})
}

func TestRunAfterBoth(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).RunAfterBoth(completable.SupplyAsync(func() string {
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
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).RunAfterBothAsync(completable.SupplyAsync(func() string {
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
}

func TestApplyEither(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ApplyToEither(completable.SupplyAsync(func() string {
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
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).ApplyToEitherAsync(completable.SupplyAsync(func() string {
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
}

func TestAcceptEither(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).AcceptEither(completable.SupplyAsync(func() string {
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
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).AcceptEitherAsync(completable.SupplyAsync(func() string {
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
}

func TestRunEither(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).RunAfterEither(completable.SupplyAsync(func() string {
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
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
			time.Sleep(1 * time.Second)
			return "Hello"
		}).RunAfterEitherAsync(completable.SupplyAsync(func() string {
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
}

func TestExceptionally(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
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
}

func TestWhenComplete(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
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
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := ""
		cf := completable.SupplyAsync(func() string {
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
}

func TestHandle(t *testing.T) {
	t.Run("sync", func(t *testing.T) {
		now := time.Now()
		ret := 0
		cf := completable.SupplyAsync(func() string {
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
	t.Run("async", func(t *testing.T) {
		now := time.Now()
		ret := 0
		cf := completable.SupplyAsync(func() string {
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
}
