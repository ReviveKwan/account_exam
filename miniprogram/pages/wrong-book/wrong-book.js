const { API } = require('../../utils/request')

const CN_NUMS = ['','一','二','三','四','五','六','七','八','九','十','十一','十二','十三','十四','十五','十六','十七','十八']

Page({
  data: {
    summary: null,
    loading: true
  },

  onShow() {
    this.loadData()
  },

  async loadData() {
    try {
      const data = await API.getWrongBook()
      if (data.chapters) {
        data.chapters = data.chapters.map(ch => ({
          ...ch,
          cnNo: CN_NUMS[ch.chapter_no] || ch.chapter_no
        }))
      }
      this.setData({ summary: data, loading: false })
    } catch (err) {
      console.error('加载错题本失败:', err)
      this.setData({ loading: false })
    }
  },

  practiceWrong() {
    wx.showToast({ title: '即将开放', icon: 'none' })
  }
})
