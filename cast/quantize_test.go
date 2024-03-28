package cast_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/wormbks/asciinema-edit/cast"
)

var _ = Describe("Quantize", func() {
	Describe("parameter validation", func() {
		var data *cast.Cast

		BeforeEach(func() {
			data = &cast.Cast{
				EventStream: []*cast.Event{
					{},
				},
			}
		})

		Context("with nil cast", func() {
			It("fails", func() {
				err := cast.Quantize(nil, nil)
				Expect(err).ToNot(Succeed())
			})
		})

		Context("with an empty event stream", func() {
			JustBeforeEach(func() {
				data = &cast.Cast{
					EventStream: []*cast.Event{},
				}
			})

			It("fails", func() {
				err := cast.Quantize(data, nil)
				Expect(err).ToNot(Succeed())
			})
		})

		Context("with a nil range list", func() {
			It("fails", func() {
				err := cast.Quantize(data, nil)
				Expect(err).ToNot(Succeed())
			})
		})

		Context("with an empty range list", func() {
			It("fails", func() {
				err := cast.Quantize(data, []cast.QuantizeRange{})
				Expect(err).ToNot(Succeed())
			})
		})
	})

	Describe("RangeOverlaps", func() {
		var qRange *cast.QuantizeRange

		BeforeEach(func() {
			qRange = &cast.QuantizeRange{
				From: 1,
				To:   2,
			}
		})

		It("doesnt overlap if no in another range", func() {
			Expect(qRange.RangeOverlaps(cast.QuantizeRange{
				From: 30,
				To:   40,
			})).ToNot(BeTrue())
		})

		It("overlaps if from in another range", func() {
			Expect(qRange.RangeOverlaps(cast.QuantizeRange{
				From: 1.5,
				To:   3,
			})).To(BeTrue())
		})

		It("overlaps if to in another range", func() {
			Expect(qRange.RangeOverlaps(cast.QuantizeRange{
				From: 0.9,
				To:   1.5,
			})).To(BeTrue())
		})
	})

	Describe("InRange", func() {
		var qRange *cast.QuantizeRange

		BeforeEach(func() {
			qRange = &cast.QuantizeRange{
				From: 1,
				To:   2,
			}
		})

		It("in range if `from <= x < to`", func() {
			Expect(qRange.InRange(1.5)).To(BeTrue())
		})

		It("in range if `x == from`", func() {
			Expect(qRange.InRange(1)).To(BeTrue())
		})

		It("not in range if `x == to`", func() {
			Expect(qRange.InRange(2)).ToNot(BeTrue())
		})

		It("not in range if `x > to`", func() {
			Expect(qRange.InRange(2.1)).ToNot(BeTrue())
		})

		It("not in range if `x < from`", func() {
			Expect(qRange.InRange(0.9)).ToNot(BeTrue())
		})
	})

	Context("having ranges specified", func() {
		var (
			data                     *cast.Cast
			event1, event2, event5   *cast.Event
			event9, event10, event11 *cast.Event
			err                      error
		)

		BeforeEach(func() {
			event1 = &cast.Event{Time: 1}
			event2 = &cast.Event{Time: 2}
			event5 = &cast.Event{Time: 5}
			event9 = &cast.Event{Time: 9}
			event10 = &cast.Event{Time: 10}
			event11 = &cast.Event{Time: 11}

			data = &cast.Cast{
				EventStream: []*cast.Event{
					event1,
					event2,
					event5,
					event9,
					event10,
					event11,
				},
			}
		})

		Context("cuts down delays with a single range", func() {
			var ranges []cast.QuantizeRange

			JustBeforeEach(func() {
				ranges = []cast.QuantizeRange{{2, 6}}
				err = cast.Quantize(data, ranges)
				Expect(err).To(Succeed())
			})

			It("modifies the timestamps accordingly", func() {
				Expect(event1.Time).To(Equal(float64(1)),
					"first")
				Expect(event2.Time).To(Equal(float64(2)),
					"second")
				Expect(event5.Time).To(Equal(float64(4)),
					"third")
				Expect(event9.Time).To(Equal(float64(6)),
					"fourth")
				Expect(event10.Time).To(Equal(float64(7)),
					"fifth")
				Expect(event11.Time).To(Equal(float64(8)),
					"sixth")
			})
		})
	})
})
