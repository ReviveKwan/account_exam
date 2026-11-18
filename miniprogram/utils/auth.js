// 微信登录工具
const { API } = require('./request')

function login() {
  return new Promise((resolve, reject) => {
    wx.login({
      success: res => {
        if (res.code) {
          API.login(res.code).then(data => {
            wx.setStorageSync('token', data.token)
            wx.setStorageSync('userInfo', data.user)
            const app = getApp()
            app.globalData.token = data.token
            app.globalData.userInfo = data.user
            resolve(data)
          }).catch(reject)
        } else {
          reject(new Error('wx.login 失败'))
        }
      },
      fail: reject
    })
  })
}

function guestLogin() {
  return new Promise((resolve, reject) => {
    API.guestLogin().then(data => {
      wx.setStorageSync('token', data.token)
      wx.setStorageSync('userInfo', data.user)
      const app = getApp()
      app.globalData.token = data.token
      app.globalData.userInfo = data.user
      resolve(data)
    }).catch(reject)
  })
}

function isLoggedIn() {
  return !!wx.getStorageSync('token')
}

function logout() {
  wx.removeStorageSync('token')
  wx.removeStorageSync('userInfo')
  const app = getApp()
  app.globalData.token = ''
  app.globalData.userInfo = null
}

module.exports = { login, guestLogin, isLoggedIn, logout }
