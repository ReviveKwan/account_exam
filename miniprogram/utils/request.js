const app = getApp()

// 正式环境仅走云托管 callContainer
// 注意：callContainer timeout 最大 15 秒，超过会无效
const REQUEST_MODE = 'cloud'
const CALL_TIMEOUT_MS = 14000

function request(options) {
  const path = options.path || '/'
  const method = options.method || 'GET'
  const data = options.data || {}
  const timeout = Math.min(options.timeout || CALL_TIMEOUT_MS, 15000)

  if (REQUEST_MODE === 'local') {
    return httpRequest((app.globalData.apiBase || '') + path, method, data)
  }

  return callContainer(path, method, data, timeout)
}

function callContainer(path, method, data, timeout) {
  const serviceName = app.globalData.serviceName || 'accounting-exam-v1'
  const header = {
    'content-type': 'application/json',
    'X-WX-SERVICE': serviceName
  }
  const token = wx.getStorageSync('token')
  if (token) header['Authorization'] = 'Bearer ' + token

  console.log('callContainer start', method, path)
  return app.ensureCloud().then(cloud => {
    return cloud.callContainer({
      path,
      method,
      header,
      data,
      timeout: timeout || CALL_TIMEOUT_MS
    })
  }).then(res => {
    console.log('callContainer done', method, path, 'status=', res && res.statusCode, 'service=', serviceName)
    if (res && res.statusCode && res.statusCode >= 400) {
      const err = (res.data && (res.data.msg || res.data.message)) || ('HTTP ' + res.statusCode)
      return Promise.reject({ code: res.statusCode, msg: err, data: res.data })
    }
    let body = res && res.data !== undefined ? res.data : res
    if (typeof body === 'string') {
      try {
        body = JSON.parse(body)
      } catch (e) {
        return Promise.reject({ code: -1, msg: '响应解析失败', data: body })
      }
    }
    return handleResponse(body)
  }).catch(err => {
    console.error('callContainer fail', err, '| service=', serviceName, '| env=', app.globalData.envId, '| path=', path)
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
      timeout: CALL_TIMEOUT_MS,
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

function errMsg(err) {
  if (!err) return '网络异常'
  if (typeof err === 'string') return err
  return err.msg || err.errMsg || err.message || '请求失败'
}

const API = {
  login: (code) => request({ path: '/api/v1/auth/login', method: 'POST', data: { code } }),
  guestLogin: () => request({ path: '/api/v1/auth/guest', method: 'POST', data: {} }),
  listChapters: () => request({ path: '/api/v1/chapters' }),
  getExamDetail: (id) => request({
    path: '/api/v1/exams/' + encodeURIComponent(id)
  }),
  getQuestion: (id) => request({
    path: '/api/v1/questions/' + encodeURIComponent(id)
  }),
  startExam: (id) => request({
    path: '/api/v1/exams/' + encodeURIComponent(id) + '/start',
    method: 'POST',
    data: {}
  }),
  resumeExam: (id) => request({
    path: '/api/v1/exams/' + encodeURIComponent(id) + '/resume'
  }),
  saveProgress: (id, data) => request({
    path: '/api/v1/exams/' + encodeURIComponent(id) + '/progress',
    method: 'PUT',
    data
  }),
  submitExam: (id, data) => request({
    path: '/api/v1/exams/' + encodeURIComponent(id) + '/submit',
    method: 'POST',
    data
  }),
  getRecord: (id) => request({ path: '/api/v1/records/' + encodeURIComponent(id) }),
  listRecords: (limit) => request({
    path: '/api/v1/records',
    data: { limit: limit || 20 }
  }),
  getWrongBook: () => request({ path: '/api/v1/wrong-questions' }),
  markCorrect: (questionId) => request({
    path: '/api/v1/wrong-questions/practice',
    method: 'POST',
    data: { question_id: questionId }
  }),
}

module.exports = { request, API, errMsg }
