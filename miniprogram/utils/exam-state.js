// 考试状态管理 - 断点续答
const STORAGE_KEY = 'exam_state'

function saveState(examId, state) {
  const all = wx.getStorageSync(STORAGE_KEY) || {}
  all[examId] = {
    ...state,
    savedAt: Date.now()
  }
  wx.setStorageSync(STORAGE_KEY, all)
}

function getState(examId) {
  const all = wx.getStorageSync(STORAGE_KEY) || {}
  return all[examId] || null
}

function clearState(examId) {
  const all = wx.getStorageSync(STORAGE_KEY) || {}
  delete all[examId]
  wx.setStorageSync(STORAGE_KEY, all)
}

// 格式化时间 mm:ss
function formatTime(seconds) {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

module.exports = { saveState, getState, clearState, formatTime }
