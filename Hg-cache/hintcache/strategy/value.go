package strategy

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}
