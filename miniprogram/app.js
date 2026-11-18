App({
  globalData: {
    token: '',
    userInfo: null,
    envId: 'wechat-app-d6gjceiz966031f7f',
    // 与控制台服务名完全一致
    serviceName: 'accounting-exam',
    // callContainer 异常时自动降级到公网（开发联调用）
    publicBase: 'https://accounting-exam-349296-12-1534378013.sh.run.tcloudbase.com',
    cloud: null,
    cloudReady: null
  },

  onLaunch() {
    const token = wx.getStorageSync('token')
    if (token) {
      this.globalData.token = token
      this.globalData.userInfo = wx.getStorageSync('userInfo') || null
    }
    this.globalData.cloudReady = this.initCloud()
  },

  // 按微信文档：先 new Cloud + await init，再 callContainer
  initCloud() {
    if (!wx.cloud) {
      console.error('请使用基础库 2.23.0+ 以使用云托管 callContainer')
      return Promise.reject(new Error('wx.cloud unavailable'))
    }
    const cloud = new wx.cloud.Cloud({
      resourceEnv: this.globalData.envId
    })
    this.globalData.cloud = cloud
    return cloud.init().then(() => {
      console.log('cloud init ok', this.globalData.envId)
      return cloud
    }).catch(err => {
      console.error('cloud init fail', err)
      throw err
    })
  },

  ensureCloud() {
    if (this.globalData.cloud) {
      return Promise.resolve(this.globalData.cloud)
    }
    if (this.globalData.cloudReady) {
      return this.globalData.cloudReady
    }
    this.globalData.cloudReady = this.initCloud()
    return this.globalData.cloudReady
  }
})
