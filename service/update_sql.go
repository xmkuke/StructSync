package service

type TimingType int

type UpdateSQL struct {
	Table  string
	Column string
	Type   AlterType
	Timing TimingType
	SQL    string
}
