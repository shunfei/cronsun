import Vue from 'vue';
import Vuex from 'vuex';

Vue.use(Vuex);

const store = new Vuex.Store({
  state: {
    enabledAuth: false,
    user: {
      email: '',
      role: 0
    },
    nodes: {},
    dropdownNodes: []
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
    },

    nodes: function (state) {
      return state.nodes;
    },

    getHostnameByID: function (state) {
      return (id) => {
        return state.nodes[id] ? state.nodes[id].hostname : id;
      }
    },

    getNodeByID: function (state) {
      return (id) => {
        return state.nodes[id]
      }
    },

    dropdownNodes: function (state) {
      return state.dropdownNodes;
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
    },

    setNodes: function (state, nodes) {
      state.nodes = nodes;
      var dn = []
      for (var i in nodes) {
        dn.push({value: nodes[i].id, name: nodes[i].hostname || nodes[i].id + '(need to upgrade)'})
      }
      state.dropdownNodes = dn;
    }
  }
})

export default store
