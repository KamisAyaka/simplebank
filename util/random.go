package util

import "math/rand"

// Go 1.20+ 会自动为 math/rand 的全局随机源播种，
// 不需要再手动调用 rand.Seed。
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(b)
}

func RandomOwner() string {
	return RandomString(6)
}

func RandomMoney() int64 {
	return RandomInt(0, 10000)
}

func RandomPositiveMoney() int64 {
	return RandomInt(1, 10000)
}

func RandomEntryAmount() int64 {
	amount := RandomInt(1, 10000)
	if rand.Intn(2) == 0 {
		return -amount
	}
	return amount
}

func RandomBool() bool {
	return rand.Intn(2) == 0
}

func RandomCurrency() string {
	return RandomString(3)
}
