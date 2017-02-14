<template>
  <div style="text-align: center;">
    <div class="ui icon buttons">
      <router-link :to="pageURL(startPage-1)" class="ui button" :class="{disabled: startPage<=1}"><i class="angle left icon"></i></router-link>
      <router-link :to="pageURL(startPage + n - 1)" v-for="n in pageBtnNum" class="ui button" :class="{blue: startPage+n-1 == _current}">{{startPage + n-1}}</router-link>
      <router-link :to="pageURL(startPage+length)" class="ui button" :class="{disabled: startPage+length>total}"><i class="angle right icon"></i></router-link>
    </div>
  </div>
</template>

<script>
export default {
  name: 'pager',
  props: ['total', 'length', 'pageVar'],
  data: function(){
    return {
      _pagevar: '',
      _current: 1,
    }
  },
  created: function(){
    this._pagevar = this.pageVar || 'page';
    this._current = this.$route.query[this._pagevar] || 1;
  },

  mounted: function(){
    console.log('mounted');
  },

  methods: {
    pageURL: function(page){
      return this.url + this._pagevar + '=' + page;
    }
  },

  watch: {
    '$route': function(){
      this._current = this.$route.query[this._pagevar] || 1;
    }
  },

  computed: {
    pageBtnNum: function(){
      console.log('pageBtnNum');
      var remainingPage = this.total - this.startPage;
      return remainingPage <= this.length ? this.total - this.startPage + 1 : this.length;
    },

    startPage: function(){
      console.log('startPage');
      return Math.floor((this._current-1)/this.length) * this.length+1;
    },

    url: function(){
      console.log('url');
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