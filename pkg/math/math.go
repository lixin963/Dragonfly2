/*
 *     Copyright 2022 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package math

func MaxInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func MaxInt32(a, b int32) int32 {
	if a > b {
		return a
	}

	return b
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}

	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func MinInt32(a, b int32) int32 {
	if a < b {
		return a
	}

	return b
}

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}

	return b
}
