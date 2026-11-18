Component({
  properties: {
    questions: { type: Array, value: [] },
    answers: { type: Object, value: {} },
    currentIndex: { type: Number, value: 0 }
  },

  data: {
    singleQuestions: [],
    multiQuestions: [],
    answeredCount: 0,
    unansweredCount: 0,
    markedCount: 0
  },

  observers: {
    'questions, answers, currentIndex': function(questions, answers, currentIndex) {
      const single = []
      const multi = []
      let answered = 0
      let marked = 0

      questions.forEach((q, idx) => {
        const a = answers[q.id] || { answer: '', marked: false }
        const isAnswered = a.answer && a.answer.length > 0
        const isMarked = a.marked
        const isCurrent = idx === currentIndex

        const item = {
          index: idx,
          questionNo: q.question_no,
          isAnswered,
          isMarked,
          isCurrent
        }

        if (isAnswered) answered++
        if (isMarked) marked++

        if (q.type === 1) {
          single.push(item)
        } else {
          multi.push(item)
        }
      })

      this.setData({
        singleQuestions: single,
        multiQuestions: multi,
        answeredCount: answered,
        unansweredCount: questions.length - answered,
        markedCount: marked
      })
    }
  },

  methods: {
    onGoTo(e) {
      this.triggerEvent('goto', { index: e.currentTarget.dataset.index })
    },

    onClose() {
      this.triggerEvent('close')
    },

    onSubmit() {
      this.triggerEvent('submit')
    },

    onContinue() {
      this.triggerEvent('close')
    }
  }
})
