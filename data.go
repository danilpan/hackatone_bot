package finbot

type SignState int

const (
	StateEmail SignState = iota
	StateTel
	StateCourse
)

type CourseSign struct {
	State     SignState // 0 - email, 1 - tel, 2 - course
	Name      string
	Email     string
	Telephone string
	Course    string
}
