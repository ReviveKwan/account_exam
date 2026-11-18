Component({
  properties: {
    total: { type: Number, value: 0 },
    answers: { type: Object, value: {} }
  },

  data: {
    answeredCount: 0,
    unansweredCount: 0,
    markedCount: 0
  },

  observers: {
    'total, answers': function(total, answers) {
      let answered = 0
      let marked = 0
      for (let key in answers) {
        const a = answers[key]
        if (a.answer && a.answer.length > 0) answered++
        if (a.marked) marked++
      }
      this.setData({
        answeredCount: answered,
        unansweredCount: total - answered,
        markedCount: marked
      })
    }
  },

  methods: {
    onConfirm() {
      this.triggerEvent('confirm')
    },
    onCancel() {
      this.triggerEvent('cancel')
    }
  }
})
