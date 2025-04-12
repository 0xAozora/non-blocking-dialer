package nb_dialer

import "testing"

func TestDialer(t *testing.T) {

	dialer := &Dialer{}

	dialer.Dial("1.1.1.1", "80")

}
