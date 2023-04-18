package finbot

type SignState int

const (
	StateTel SignState = iota
	StateRegistered
	StateBuilding
	StateGuestAdd
	StateNumberChangeState
)

type CourseSign struct {
	State     SignState // 0 - email, 1 - tel, 2 - course
	Telephone string
	Building  int
	NumberId  int
}
