<style scope>
  .clearfix:after {content:""; clear:both; display:table;}
</style>
<template>
  <div>
    <div class="clearfix" style="margin-bottom: 20px;">
      <router-link class="ui left floated button" to="/job">查看任务列表</router-link>
      <button class="ui left floated icon button" v-on:click="refresh"><i class="refresh icon"></i></button>
      <router-link class="ui right floated primary button" to="/job/create"><i class="add to calendar icon"></i> 新任务</router-link>
    </div>
    <form class="ui form" v-bind:class="{loading:loading}" v-on:submit.prevent>
      <div class="field">
        <label>任务 ID</label>
        <input type="text" ref="ids" v-model:value="ids" placeholder="多个 ID 使用英文逗号分隔"/>
      </div>
      <div class="field">
        <label>选择分组</label>
        <Dropdown title="选择分组" v-bind:items="prefetchs.groups" v-on:change="changeGroup" :selected="groups" :multiple="true"/>
      </div>
      <div class="field">
        <label>选择节点</label>
        <Dropdown title="选择节点" v-bind:items="prefetchs.nodes" v-on:change="changeNodes" :selected="nodes" :multiple="true"/>
      </div>
      <div class="field">
        <button class="fluid ui button" type="button" v-on:click="submit">查询</button>
      </div>
    </form>
    <table class="ui hover blue table" v-if="executings.length > 0">
      <thead>
        <tr>
          <th class="center aligned">任务ID</th>
          <th width="200px" class="center aligned">分组</th>
          <th class="center aligned">节点</th>
          <th class="center aligned">进程ID</th>
          <th class="center aligned">开始时间</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(proc, index) in executings">
          <td class="center aligned"><router-link :to="'/job/edit/'+proc.group+'/'+proc.jobId">{{proc.jobId}}</router-link></td>
          <td class="center aligned">{{proc.group}}</td>
          <td class="center aligned">{{proc.nodeId}}</td>
          <td class="center aligned">{{proc.id}}</td>
          <td class="center aligned">{{proc.time}}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
import Dropdown from './basic/Dropdown.vue';
import {split} from '../libraries/functions';

export default {
  name: 'job-executing',
  data(){
    return {
      prefetchs: {groups: [], nodes: []},
      loading: false,
      groups: [],
      ids: '',
      nodes: [],
      executings: []
    }
  },
  
  mounted(){
    var vm = this;
    this.groups = split(this.$route.query.groups, ',');
    this.nodes = split(this.$route.query.nodes, ',');
    this.ids = this.$route.query.ids || '';

    this.$rest.GET('job/groups').onsucceed(200, (resp)=>{
      !resp.includes('default') && resp.unshift('default');
      vm.prefetchs.groups = resp;
      this.fetchList(this.buildQuery());
    }).do();

    this.$rest.GET('nodes').onsucceed(200, (resp)=>{
      for (var i in resp) {
        vm.prefetchs.nodes.push(resp[i].id);
      }
    }).do();
  },

  watch: {
    '$route': function(){
      this.groups = split(this.$route.query.groups, ',');
      this.nodes = split(this.$route.query.nodes, ',');
      this.ids = this.$route.query.ids || '';
      this.fetchList(this.buildQuery());
    }
  },

  methods: {
    changeGroup(val, text){
      this.groups = split(val, ',');
    },

    changeNodes(val){
      this.nodes = split(val, ',');
    },

    submit(){
      this.$router.push('/job/executing?'+this.buildQuery());
    },

    buildQuery(){
      var params = [];
      if (this.groups && this.groups.length > 0) params.push('groups='+this.groups.join(','));
      if (this.nodes && this.nodes.length > 0) params.push('nodes='+this.nodes.join(','));
      if (this.ids) params.push('ids='+this.ids);
      return params.join('&');
    },

    fetchList(query){
      var vm = this;
      this.loading = true;
      this.$rest.GET('job/executing?'+query).
      onsucceed(200, (resp)=>{
        vm.executings = resp;
        vm.$nextTick(()=>{
          $(vm.$el).find('table .ui.dropdown').dropdown();
        });
      }).
      onend(()=>{vm.loading = false}).
      do();
    },

    refresh(){
      this.fetchList(this.buildQuery());
    }
  },

  components: {
    Dropdown
  }
}
</script>