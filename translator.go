/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

type Translator interface {
	ParseFile(f string)
	Parse(buf []byte)
	Get(str string) string
	GetN(str, plural string, n int) string
	GetC(str, ctx string) string
	GetNC(str, plural string, n int, ctx string) string
}
