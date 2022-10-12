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

package completable

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

func GetAny(ctx context.Context, cfs ...CompletionStage) (cs CompletionStage, err error) {
	if len(cfs) == 0 {
		return nil, errors.New("CompletionStage size is 0. ")
	}
	chs, err := selectChannels(ctx, cfs...)
	i, cs := selectCompletionStage(context.Background(), chs...)
	if i == len(cfs) {
		return nil, errors.New("Context Done. ")
	}
	return cs, nil
}

func selectChannels(ctx context.Context, vhs ...CompletionStage) (channels []chan CompletionStage, err error) {
	channels = make([]chan CompletionStage, len(vhs))
	for i, c := range vhs {
		ch := make(chan CompletionStage)
		cs := c
		channels[i] = ch

		if joinable, ok := cs.(Joinable); ok {
			go func(ep *error) {
				origin := joinable.JoinCompletionStage(ctx)
				// Get maybe panic, ignore it and return the CompletionStage
				defer func(ep *error, ret CompletionStage) {
					if o := recover(); o != nil {
						if e, ok := o.(error); ok {
							*ep = e
						} else {
							*ep = fmt.Errorf("AnyOf panic: %v . ", o)
						}
					}
					ch <- origin
				}(ep, origin)
				origin.Get(nil)
			}(&err)
		}

	}
	return channels, nil
}

func selectCompletionStage(ctx context.Context, vhs ...chan CompletionStage) (int, CompletionStage) {
	size := len(vhs)
	if ctx != nil {
		size++
	}
	selectCases := make([]reflect.SelectCase, size)
	if ctx != nil {
		selectCases[len(vhs)] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		}
	}
	for i, vh := range vhs {
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(vh),
		}
	}
	index, value, _ := reflect.Select(selectCases)
	if index == len(vhs) {
		return index, nil
	}
	return index, value.Interface().(CompletionStage)
}
