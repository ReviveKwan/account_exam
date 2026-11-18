const app = getApp()

// 'public' = 直接走公网（当前联调推荐）
// 'cloud'  = 优先 callContainer，失败再降级公网
// 'local'  = localhost
const REQUEST_MODE = 'public'

function request(options) {
  const path = options.path || '/'
  const method = options.method || 'GET'
  const data = options.data || {}
  const publicBase = (app.globalData && app.globalData.publicBase) || ''

  if (REQUEST_MODE === 'local') {
    return httpRequest((app.globalData.apiBase || '') + path, method, data)
  }

  if (REQUEST_MODE === 'public') {
    if (!publicBase) {
      return Promise.reject(new Error('未配置 publicBase 公网域名'))
    }
    return httpRequest(publicBase + path, method, data)
  }

  // cloud 模式
  return callContainer(path, method, data).catch(err => {
    const msg = (err && (err.errMsg || err.message || String(err))) || ''
    if (publicBase && /SERVICE_VERSION_NOT_FOUND|version not found|callContainer/i.test(msg)) {
      console.warn('callContainer 失败，降级公网:', msg)
      return httpRequest(publicBase + path, method, data)
    }
    return Promise.reject(err)
  })
}

function callContainer(path, method, data) {
  const serviceName = app.globalData.serviceName || 'accounting-exam'
  const header = {
    'content-type': 'application/json',
    'X-WX-SERVICE': serviceName
  }
  const token = wx.getStorageSync('token')
  if (token) header['Authorization'] = 'Bearer ' + token

  return app.ensureCloud().then(cloud => {
    return cloud.callContainer({
      path,
      method,
      header,
      data,
      timeout: 15000
    })
  }).then(res => {
    console.log('callContainer', method, path, 'status=', res.statusCode, 'service=', serviceName)
    const body = res && res.data !== undefined ? res.data : res
    return handleResponse(body)
  }).catch(err => {
    console.error('callContainer fail', err, '| service=', serviceName, '| env=', app.globalData.envId)
    return Promise.reject(err)
  })
}

function httpRequest(url, method, data) {
  const header = { 'Content-Type': 'application/json' }
  const token = wx.getStorageSync('token')
  if (token) header['Authorization'] = 'Bearer ' + token

  return new Promise((resolve, reject) => {
    wx.request({
      url,
      method,
      header,
      data,
      success: res => {
        handleResponse(res.data).then(resolve).catch(reject)
      },
      fail: reject
    })
  })
}

function handleResponse(data) {
  return new Promise((resolve, reject) => {
    if (!data || typeof data !== 'object') {
      reject(data || new Error('empty response'))
      return
    }
    if (data.code === 0) {
      resolve(data.data)
    } else if (data.code === 4003) {
      wx.removeStorageSync('token')
      wx.redirectTo({ url: '/pages/login/login' })
      reject(data)
    } else {
      reject(data)
    }
  })
}

const API = {
  login: (code) => request({ path: '/api/v1/auth/login', method: 'POST', data: { code } }),
  guestLogin: () => request({ path: '/api/v1/auth/guest', method: 'POST', data: {} }),
  listChapters: () => request({ path: '/api/v1/chapters' }),
  getExamDetail: (id, mode) => request({ path: `/api/v1/exams/${id}?mode=${mode || 1}` }),
  startExam: (id) => request({ path: `/api/v1/exams/${id}/start`, method: 'POST', data: {} }),
  resumeExam: (id) => request({ path: `/api/v1/exams/${id}/resume` }),
  saveProgress: (id, data) => request({ path: `/api/v1/exams/${id}/progress`, method: 'PUT', data }),
  submitExam: (id, data) => request({ path: `/api/v1/exams/${id}/submit`, method: 'POST', data }),
  getRecord: (id) => request({ path: `/api/v1/records/${id}` }),
  listRecords: (limit) => request({ path: `/api/v1/records?limit=${limit || 20}` }),
  getWrongBook: () => request({ path: '/api/v1/wrong-questions' }),
  markCorrect: (questionId) => request({ path: '/api/v1/wrong-questions/practice', method: 'POST', data: { question_id: questionId } }),
}

module.exports = { request, API }
