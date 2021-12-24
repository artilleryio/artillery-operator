/*
 * Copyright (c) 2022.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0.
 *
 * If a copy of the MPL was not distributed with
 * this file, You can obtain one at
 *
 *     http://mozilla.org/MPL/2.0/
 */

package controllers

// vusers.created_by_name.Access the / route: .................. 150
// vusers.created.total: ....................................... 150
// vusers.completed: ...........................................150
// vusers.session_length:
//   min: ...................................................... 40.3
//   max: ...................................................... 157.5
//   median: ................................................... 45.2
//   p95: ...................................................... 55.2
//   p99: ...................................................... 156
// http.request_rate: .......................................... 3/sec
// http.requests: ..............................................150
// http.codes.200: ............................................. 150
// http.responses: ............................................. 150
// http.response_time:
//   min: ...................................................... 16
//   max: ...................................................... 33
//   median: ................................................... 19.1
//   p95: ...................................................... 22.9
//   p99: ...................................................... 25.8

// var metricsPatterns = map[string]*regexp.Regexp{
// 	"scenarios.created":   regexp.MustCompile("vusers.created.total: ....................................... 150"),
// 	"scenarios_completed": regexp.MustCompile("vusers.completed: ...........................................150"),
// 	"requests_completed":  regexp.MustCompile("http.requests: ..............................................150"),
// 	"latency_min":         regexp.MustCompile("http.response_time.*min: ...................................................... 16"),
// 	"latency_max":         regexp.MustCompile("http.response_time:.*max: ...................................................... 33"),
// 	"latency_median":      regexp.MustCompile("http.response_time:.*median  ................................................... 19.1"),
// 	"latency_p95":         regexp.MustCompile("http.response_time:.*p95: ...................................................... 22.9"),
// 	"response_200":        regexp.MustCompile("http.response_time:.*p99: ...................................................... 25.8"),
// 	"rps_mean":            regexp.MustCompile("http.request_rate: .......................................... 3/sec"),
// 	"rps_count":           regexp.MustCompile("http.responses: ............................................. 150"),
// }
