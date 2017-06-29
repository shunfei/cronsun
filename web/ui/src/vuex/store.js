import Vue from 'vue';
import Vuex from 'vuex';

Vue.use(Vuex);

export default new Vuex.Store({
  state: {
    enabledAuth: false,
    user: {
      email: '',
      role: 0
    }
  },

  getters: {
    email: function (state) {
      return state.user.email;
    },

    role: function (state) {
      return state.user.role;
    },

    enabledAuth: function (state) {
      return state.enabledAuth;
    }
  },

  mutations: {
    setEmail: function (state, email) {
      state.user.email = email;
    },

    setRole: function (state, role) {
      state.user.role = role;
    },

    enabledAuth: function (state, enabledAuth) {
      state.enabledAuth = enabledAuth;
    }
  }
})
