<template>
  <div v-if="error != ''" class="ui negative message">
    <div class="header"><i class="attention icon"></i> {{error}}</div>
  </div>
  <form v-else class="ui form" v-bind:class="{loading:loading}" v-on:submit.prevent>
    <h3 class="ui header">{{action == 'CREATE' ? '添加' : '更新'}}任务&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
      <div class="ui toggle checkbox">
        <input type="checkbox" class="hidden" v-bind:checked="!job.pause">
        <label v-bind:style="{color: (job.pause?'red':'green')+' !important'}">{{job.pause ? '任务已暂停' : '开启'}}</label>
      </div>
      <em v-if="job.id">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; ID# {{job.id}}</em>
    </h3>
    <div class="two fields">
      <div class="field">
        <label>任务名称</label>
        <input type="text" ref="name" v-bind:value="job.name" v-on:input="updateValue($event.target.value)" placeholder="任务名称">
      </div>
      <div class="field">
        <label>任务分组</label>
        <Dropdown title="选择分组" v-bind:items="groups" v-bind:selected="job.group" v-on:change="changeGroup"/>
      </div>
    </div>
    <div class="fields">
      <div class="twelve wide field">
        <label>任务脚本</label>
        <input type="text" v-model="job.cmd" placeholder="任务脚本">
      </div>
      <div class="four wide field">
        <label>用户(可选)</label>
        <input type="text" v-model="job.user" placeholder="指定执行脚本的用户">
      </div>
    </div>
    <div class="field">
      <span v-if="!job.rules || job.rules.length == 0"><i class="warning circle icon"></i>当前任务没有定时器，点击下面按钮来添加定时器</span>
    </div>
    <JobEditRule v-for="(rule, index) in job.rules" :key="rule.id" v-bind:rule="rule" :index="index" v-on:remove="removeRule" v-on:change="changeRule"/>
    <div class="two fields">
      <div class="field">
        <button class="fluid ui button" v-on:click="addNewTimer" type="button"><i class="history icon"></i> 添加定时器</button>
      </div>
      <div class="field">
        <button class="fluid blue ui button" type="button" v-on:click="submit"><i class="upload icon"></i> 保存任务</button>
      </div>
    </div>
  </form>
</template>

<script>
import JobEditRule from './JobEditRule.vue';
import Dropdown from './basic/Dropdown.vue';

export default {
  name: 'job-edit',
  data: function(){
      return {
        action: 'CREATE',
        groups: [],
        loading: false,
        job: {
          id: '',
          name:  '',
          group: 'default',
          user: '',
          cmd: '',
          pause: false,
          rules: []
        },
        error: ''
      }
  },

  methods: {
    updateValue: function(v){
      var tv = v.replace(/[\*\/]/g, '');
      this.job.name = tv;
      if (tv !== v) {
        this.$refs.name.value = tv;
      }
    },

    addNewTimer: function(){
      if (!this.job.rules) this.job.rules = [];
      this.job.rules.push({id: this.newRandomRuleId()});
    },

    changeGroup: function(val, text){
      this.job.group = val;
    },

    removeRule: function(index){
      this.job.rules.splice(index, 1);
    },

    changeRule: function(index, key, val){
      this.job.rules[index][key] = val;
    },

    submit: function(){
      var exceptCode = this.action == 'CREATE' ? 201 : 200;
      this.loading = true;
      var vm = this;
      this.$rest.PUT('job', this.job)
        .onsucceed(exceptCode, ()=>{vm.$router.push('/job')})
        .onfailed((resp)=>{console.log(resp)})
        .onend(()=>{vm.loading=false})
        .do();
    },

    newRandomRuleId: function(){
      return 'NEW'+Math.random().toString();
    }
  },

  mounted: function(){
    var vm = this;

    if (this.$route.path.indexOf('/job/create') === 0) {
      this.action = 'CREATE';
    } else {
      this.action = 'UPDATE';
      this.$rest.GET('job/'+this.$route.params.group+'-'+this.$route.params.id).
        onsucceed(200, (resp)=>{
          vm.job = resp;
          if (vm.job.rules) {
            for (var i in vm.job.rules) {
              if (vm.job.rules[i].id.length == 0) {
                vm.job.rules[i].id = vm.newRandomRuleId();
              }
            }
          }
        }).
        onfailed((data)=>{vm.error = data.error}).
        do();
    }

    this.$rest.GET('job/groups').onsucceed(200, (resp)=>{
      !resp.includes('default') && resp.unshift('default');
      vm.groups = resp;
    }).do();

    $(this.$el).find('.checkbox').checkbox({
      onChange: function(){
        vm.job.pause = !vm.job.pause;
      }
    });

    $(this.$el).find('.dropdown').dropdown({
      allowAdditions: true,
      onChange: function(value, text, $choice){
        vm.job.group = value;
      }
    }).dropdown('set exactly', this.job.group);
  },

  components: {
    JobEditRule,
    Dropdown
  }
}
</script>