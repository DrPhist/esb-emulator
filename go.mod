module github.com/DrPhist/esb-emulator

go 1.23.4

replace github.com/DrPhist/esb-emulator/internal => ./internal

require (
	github.com/DrPhist/esb-emulator/internal v0.0.0-00010101000000-000000000000
	github.com/segmentio/kafka-go v0.4.48
)

require (
	github.com/brianvoe/gofakeit/v7 v7.3.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
)
