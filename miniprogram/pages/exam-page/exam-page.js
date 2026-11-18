const { API, errMsg } = require('../../utils/request')
const { formatTime, saveState, getState, clearState } = require('../../utils/exam-state')

function sortAnswer(str) {
  return (str || '').toUpperCase().replace(/[^A-Z]/g, '').split('').sort().join('')
}

/** 本地判题：返回 { status: 'correct'|'partial'|'wrong'|'empty', label } */
function judgeLocal(question, userAnswer) {
  const ua = sortAnswer(userAnswer)
  const ca = sortAnswer(question.answer)
  if (!ua) return { status: 'empty', label: '未作答' }
  if (question.type === 1) {
    if (ua === ca) return { status: 'correct', label: '回答正确' }
    return { status: 'wrong', label: '回答错误' }
  }
  // 多选：全对 / 少选半对 / 错选不得分
  if (ua === ca) return { status: 'correct', label: '回答正确' }
  const correctSet = {}
  ca.split('').forEach(l => { correctSet[l] = true })
  let hasWrong = false
  let correctSelected = 0
  ua.split('').forEach(l => {
    if (correctSet[l]) correctSelected++
    else hasWrong = true
  })
  if (hasWrong || correctSelected === 0) return { status: 'wrong', label: '回答错误' }
  return { status: 'partial', label: '少选半对' }
}

Page({
  data: {
    examId: null,
    mode: 1,
    chapterId: null,
    isResume: false,
    exam: null,
    questions: [],
    currentIndex: 0,
    currentQuestion: null,
    optionList: [],
    optionCount: 0,
    currentMarked: false,
    currentRevealed: false,
    currentHasAnswer: false,
    feedback: null, // { status, label, correctAnswer }
    answers: {}, // { questionId: { answer, marked, revealed } }
    recordId: null,
    remainingSeconds: 0,
    timerText: '00:00',
    showAnswerSheet: false,
    showSubmitConfirm: false,
    loading: true,
    fontScale: 'standard',
    submitted: false,
    isLastQuestion: false
  },

  timer: null,

  onLoad(options) {
    const examId = decodeURIComponent(options.examId || '')
    const mode = parseInt(options.mode || '1', 10) || 1
    const isResume = options.resume === '1'
    const chapterId = options.chapterId || null
    this.setData({ examId, mode, isResume, chapterId })
    this.loadExam(examId, mode, isResume)
  },

  onUnload() {
    this.stopTimer()
    this.saveProgress()
  },

  async loadExam(examId, mode, isResume) {
    if (!examId) {
      wx.showToast({ title: '缺少试卷ID', icon: 'none' })
      this.setData({ loading: false })
      return
    }
    try {
      wx.showLoading({ title: '加载题目...', mask: true })
      // 后端已始终返回答案/解析；真机避免 path 带 ?mode=
      const examData = await API.getExamDetail(examId)
      const exam = (examData && examData.exam) || examData || {}
      const questions = exam.questions || []
      if (!questions.length) {
        throw { msg: '题目列表为空，请稍后重试' }
      }

      let recordId = null
      let remainingSeconds = (exam.duration_minutes || 60) * 60

      if (isResume) {
        const recordData = await API.resumeExam(examId)
        recordId = recordData.record.id
        remainingSeconds = recordData.record.remaining_seconds
      } else {
        const recordData = await API.startExam(examId)
        recordId = recordData.record.id
      }

      const savedState = getState(examId)
      let answers = {}
      let startIndex = 0
      if (savedState && savedState.answers) {
        answers = savedState.answers
        if (typeof savedState.currentIndex === 'number') {
          startIndex = Math.min(savedState.currentIndex, Math.max(questions.length - 1, 0))
        }
      }

      this.setData({
        exam,
        questions,
        recordId,
        remainingSeconds,
        answers,
        currentIndex: startIndex,
        currentQuestion: questions[startIndex] || null,
        timerText: formatTime(remainingSeconds),
        loading: false,
        isLastQuestion: questions.length > 0 && startIndex === questions.length - 1
      })
      this.refreshQuestionView()
      this.startTimer()
    } catch (err) {
      console.error('加载考试失败:', err)
      const msg = errMsg(err)
      wx.showModal({
        title: '题目加载失败',
        content: msg,
        showCancel: false
      })
      this.setData({ loading: false })
    } finally {
      wx.hideLoading()
    }
  },

  refreshQuestionView() {
    const q = this.data.currentQuestion
    if (!q) {
      this.setData({
        optionList: [],
        optionCount: 0,
        currentMarked: false,
        currentRevealed: false,
        currentHasAnswer: false,
        feedback: null,
        isLastQuestion: false
      })
      return
    }
    const cur = this.data.answers[q.id] || { answer: '', marked: false, revealed: false }
    const selected = cur.answer ? cur.answer.split('') : []
    const revealed = !!cur.revealed
    const correctSet = {}
    sortAnswer(q.answer).split('').forEach(l => { correctSet[l] = true })

    const keys = Object.keys(q.options || {}).sort()
    const optionList = keys.map(key => {
      const isSelected = selected.indexOf(key) > -1
      let state = ''
      if (revealed) {
        if (correctSet[key]) state = 'correct'
        else if (isSelected) state = 'wrong'
      }
      return {
        key,
        val: q.options[key],
        selected: isSelected,
        state
      }
    })

    let feedback = null
    if (revealed) {
      const judged = judgeLocal(q, cur.answer)
      feedback = {
        status: judged.status,
        label: judged.label,
        correctAnswer: q.answer || ''
      }
    }

    this.setData({
      optionList,
      optionCount: optionList.length,
      currentMarked: !!cur.marked,
      currentRevealed: revealed,
      currentHasAnswer: !!(cur.answer && cur.answer.length),
      feedback,
      isLastQuestion: this.data.currentIndex >= this.data.questions.length - 1
    })
  },

  persistAnswers(answers) {
    this.setData({ answers })
    saveState(this.data.examId, {
      answers,
      currentIndex: this.data.currentIndex,
      remainingSeconds: this.data.remainingSeconds
    })
  },

  startTimer() {
    this.timer = setInterval(() => {
      let remaining = this.data.remainingSeconds - 1
      if (remaining <= 0) {
        remaining = 0
        this.stopTimer()
        wx.showToast({ title: '时间到，自动交卷', icon: 'none' })
        this.handleSubmit(true)
      }
      this.setData({
        remainingSeconds: remaining,
        timerText: formatTime(remaining)
      })
    }, 1000)
  },

  stopTimer() {
    if (this.timer) {
      clearInterval(this.timer)
      this.timer = null
    }
  },

  toggleOption(e) {
    const option = e.currentTarget.dataset.option
    const question = this.data.currentQuestion
    if (!question) return

    const existing = this.data.answers[question.id] || {}
    if (existing.revealed) {
      wx.showToast({ title: '本题已确认，不可再改', icon: 'none' })
      return
    }

    let answers = { ...this.data.answers }
    let current = { answer: existing.answer || '', marked: !!existing.marked, revealed: false }

    if (question.type === 1) {
      current.answer = option
    } else {
      let selected = current.answer ? current.answer.split('') : []
      const idx = selected.indexOf(option)
      if (idx > -1) selected.splice(idx, 1)
      else selected.push(option)
      selected.sort()
      current.answer = selected.join('')
    }

    answers[question.id] = current
    this.persistAnswers(answers)
    this.refreshQuestionView()
  },

  /** 确认本题：判对错；解析按需拉取以减小首包 */
  async confirmAnswer() {
    const question = this.data.currentQuestion
    if (!question) return
    const existing = this.data.answers[question.id] || { answer: '', marked: false, revealed: false }
    if (existing.revealed) return
    if (!existing.answer) {
      wx.showToast({ title: '请先选择答案', icon: 'none' })
      return
    }

    // 首包未带解析时，确认后再拉单题详情
    if (!question.analysis || !question.law_citation) {
      try {
        const data = await API.getQuestion(question.id)
        const full = (data && data.question) || {}
        const questions = this.data.questions.slice()
        const idx = this.data.currentIndex
        questions[idx] = {
          ...question,
          answer: full.answer || question.answer,
          analysis: full.analysis || question.analysis || '',
          law_citation: full.law_citation || question.law_citation || ''
        }
        this.setData({
          questions,
          currentQuestion: questions[idx]
        })
      } catch (err) {
        console.warn('拉取解析失败，仍展示对错:', err)
      }
    }

    const answers = { ...this.data.answers }
    answers[question.id] = { ...existing, revealed: true }
    this.persistAnswers(answers)
    this.refreshQuestionView()
  },

  toggleMark() {
    const question = this.data.currentQuestion
    if (!question) return
    let answers = { ...this.data.answers }
    let current = answers[question.id] || { answer: '', marked: false, revealed: false }
    current.marked = !current.marked
    answers[question.id] = current
    this.persistAnswers(answers)
    this.refreshQuestionView()
  },

  goToIndex(idx) {
    if (idx < 0 || idx >= this.data.questions.length) return
    this.setData({
      currentIndex: idx,
      currentQuestion: this.data.questions[idx],
      isLastQuestion: idx >= this.data.questions.length - 1,
      showAnswerSheet: false
    })
    this.refreshQuestionView()
    saveState(this.data.examId, {
      answers: this.data.answers,
      currentIndex: idx,
      remainingSeconds: this.data.remainingSeconds
    })
  },

  prevQuestion() {
    this.goToIndex(this.data.currentIndex - 1)
  },

  nextQuestion() {
    // 未确认时，点下一题先确认
    if (!this.data.currentRevealed) {
      if (!this.data.currentHasAnswer) {
        wx.showToast({ title: '请先选择答案', icon: 'none' })
        return
      }
      this.confirmAnswer()
      return
    }
    if (this.data.isLastQuestion) {
      this.showSubmit()
      return
    }
    this.goToIndex(this.data.currentIndex + 1)
  },

  goToQuestion(e) {
    const idx = parseInt(e.detail.index !== undefined ? e.detail.index : e.currentTarget.dataset.index, 10)
    this.goToIndex(idx)
  },

  toggleAnswerSheet() {
    this.setData({ showAnswerSheet: !this.data.showAnswerSheet })
  },

  showSubmit() {
    this.setData({ showSubmitConfirm: true })
  },

  hideSubmit() {
    this.setData({ showSubmitConfirm: false })
  },

  async saveProgress() {
    if (!this.data.recordId || this.data.submitted) return
    const answersArr = this.data.questions.map(q => {
      const a = this.data.answers[q.id] || { answer: '', marked: false }
      return { question_id: q.id, user_answer: a.answer || '', is_marked: a.marked || false }
    })
    try {
      await API.saveProgress(this.data.examId, {
        record_id: this.data.recordId,
        current_question_no: this.data.currentIndex + 1,
        remaining_seconds: this.data.remainingSeconds,
        answers: answersArr
      })
    } catch (err) {
      console.error('保存进度失败:', err)
    }
  },

  async handleSubmit(force) {
    if (!force) {
      this.setData({ showSubmitConfirm: false })
    }
    if (this.data.submitted) return
    this.stopTimer()

    const answersArr = this.data.questions.map(q => {
      const a = this.data.answers[q.id] || { answer: '', marked: false }
      return { question_id: q.id, user_answer: a.answer || '', is_marked: a.marked || false }
    })

    const durationSeconds = this.data.exam.duration_minutes * 60 - this.data.remainingSeconds

    wx.showLoading({ title: '判分中...' })
    try {
      const result = await API.submitExam(this.data.examId, {
        record_id: this.data.recordId,
        duration_seconds: durationSeconds,
        answers: answersArr
      })
      this.setData({ submitted: true })
      clearState(this.data.examId)
      const recordId = result.record_id || this.data.recordId
      wx.redirectTo({
        url: `/pages/exam-result/exam-result?recordId=${encodeURIComponent(recordId)}`
      })
    } catch (err) {
      console.error('提交失败:', err)
      wx.showToast({ title: '提交失败', icon: 'none' })
      this.startTimer()
    } finally {
      wx.hideLoading()
    }
  },

  toggleFontScale() {
    this.setData({
      fontScale: this.data.fontScale === 'standard' ? 'large' : 'standard'
    })
  }
})
