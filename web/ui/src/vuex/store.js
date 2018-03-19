import Vue from 'vue';
import Vuex from 'vuex';

Vue.use(Vuex);

const store = new Vuex.Store({
  state: {
    version: '',
    enabledAuth: false,
    user: {
      email: '',
      role: 0
    },
    nodes: {},
    showWithHostname: false
  },

  getters: {
    version: function (state) {
      return state.version;
    },

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

    showWithHostname: function (state) {
      return state.showWithHostname;
    },

    hostshows: function (state) {
      return (id) => _hostshows(id, state, true);
    },

    hostshowsWithoutTip: function (state) {
      return (id) => _hostshows(id, state, false);
    },

    getNodeByID: function (state) {
      return (id) => {
        return state.nodes[id]
      }
    },

    dropdownNodes: function (state) {
      var dn = [];
      var nodes = state.nodes;
      for (var i in nodes) {
        dn.push({
          value: nodes[i].id,
          name: _hostshows(nodes[i].id, state, true)
        });
      }
      return dn;
    }
  },

  mutations: {
    setVersion: function (state, v) {
      state.version = v;
    },

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
    },

    setShowWithHostname: function (state, b) {
      state.showWithHostname = b;
    }
  }
})

function _hostshows(id, state, tip) {
  if (!state.nodes[id]) {
    if (tip) id += '(node not found)';
    return id;
  }

  var show = state.showWithHostname ? state.nodes[id].hostname : state.nodes[id].ip;
  if (!show) {
    show = id
    if (tip) show += '(need to upgrade)';
  }
  return show;
}

export default store
