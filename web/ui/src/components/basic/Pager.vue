<template>
  <div style="text-align: center; margin-bottom: 1em;">
    <div class="ui icon buttons">
      <router-link :to="pageURL(_startPage-1)" class="ui button" :class="{disabled: _startPage<=1}"><i class="angle left icon"></i></router-link>
      <router-link :to="pageURL(_startPage + n - 1)" v-for="n in _pageBtnNum" class="ui button" :class="{blue: _startPage+n-1 == _current}">{{_startPage + n-1}}</router-link>
      <a class="ui button disabled">{{_current}}/{{total}}</a>
      <router-link :to="pageURL(_startPage+length)" class="ui button" :class="{disabled: _startPage+length>total}"><i class="angle right icon"></i></router-link>
    </div>
    <div class="ui action input">
      <input type="text" ref="gopage" style="width: 70px;">
      <button class="ui icon button" v-on:click="go">
        <i class="arrow right icon"></i>
      </button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'pager',
  props: {
    total: {type: Number, default: 1, required: true},
    length: {type: Number, default: 5}
  },
  data: function(){
    return {
      _pagevar: '',
      _current: 1,
      _startPage: 1,
      _pageBtnNUm: 5
    }
  },

  created: function(){
    this._pagevar = this.pageVar || 'page';
    this._current = this.$route.query[this._pagevar] || 1;
    this._startPage = Math.floor((this._current-1)/this.length) * this.length + 1;
    this._pageBtnNum = this.total - this._startPage - this.length <= 0 ? this.total - this._startPage + 1 : this.length;
  },

  methods: {
    pageURL: function(page){
      return this.url + this._pagevar + '=' + page;
    },

    go: function(){
      var page = +this.$refs.gopage.value;
      if (page < 1 || page > this.total) return;
      this.$router.push(this.pageURL(page));
    }
  },

  watch: {
    '$route': function(){
      this._current = this.$route.query[this._pagevar] || 1;
      this._startPage = Math.floor((this._current-1)/this.length) * this.length + 1;
      this._pageBtnNum = this.total - this._startPage - this.length <= 0 ? this.total - this._startPage + 1 : this.length;
    }
  },

  computed: {
    url: function(){
      var query = [];
      for (var  k in this.$route.query) {
        if (this._pagevar === k) continue;
        query.push(k+'='+this.$route.query[k]);
      }

      return this.$route.path+'?'+query.join('&') + '&';
    }
  }
}
</script>