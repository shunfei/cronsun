<template>
  <div v-if="error != ''" class="ui negative message">
    <div class="header"><i class="attention icon"></i> {{error}}</div>
  </div>
  <form v-else class="ui form" v-bind:class="{loading:loading}" v-on:submit.prevent>
    <h3 class="ui header">{{action == 'CREATE' ? '添加' : '更新'}}任务&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
      <div class="ui toggle checkbox" ref="pause">
        <input type="checkbox" class="hidden" v-bind:checked="!job.pause">
        <label v-bind:style="{color: (job.pause?'red':'green')+' !important'}">{{job.pause ? '任务已暂停' : '开启'}}</label>
      </div>
      <em v-if="job.id">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; ID# {{job.id}}</em>
    </h3>
    <div class="inline fields" ref="kind">
      <label>任务类型</label>
      <div class="field">
        <div class="ui radio checkbox">
          <input type="radio" v-model="job.kind" name="kind" value="0" tabindex="0" class="hidden"/>
          <label>普通任务</label>
        </div>
      </div>
      <div class="field">
        <div class="ui radio checkbox">
          <input type="radio" v-model="job.kind" name="kind" value="1" tabindex="0" class="hidden"/>
          <label title="同一时间只有一个任务进程在某个节点上面执行">单机单进程
            <i class="help circle link icon" data-position="top right" data-html="同一时间只有一个任务进程在某个节点上面执行" data-variation="wide"></i>
          </label>
        </div>
      </div>
      <div class="field">
        <div class="ui radio checkbox">
          <input type="radio" v-model="job.kind" name="kind" value="2" tabindex="0" class="hidden"/>
          <label>一个任务执行间隔内允许执行一次</label>
        </div>
      </div>
    </div>
    <div class="two fields">
      <div class="field">
        <label>任务名称</label>
        <input type="text" ref="name" v-bind:value="job.name" v-on:input="updateValue($event.target.value)" placeholder="任务名称">
      </div>
      <div class="field">
        <label>任务分组</label>
        <Dropdown title="选择分组" v-bind:allowAdditions="true" v-bind:items="groups" v-bind:selected="job.group" v-on:change="changeGroup"></Dropdown>
      </div>
    </div>
    <div class="fields">
      <div class="twelve wide field">
        <label>任务脚本 {{allowSuffixsTip}}</label>
        <input type="text" v-model="job.cmd" placeholder="任务脚本">
      </div>
      <div class="four wide field">
        <label>用户({{$appConfig.security.open ? '必选' : '可选'}})</label>
        <Dropdown v-if="$appConfig.security.open" title="指定执行用户" v-bind:items="$appConfig.security.users" v-bind:selected="job.user" v-on:change="changeUser"></Dropdown>
        <input v-else type="text" v-model="job.user" placeholder="指定执行用户">
      </div>
    </div>
    <div class="two fields">
      <div class="field">
        <label>超时设置（单位“秒”，0 表示不限制）</label>
        <input type="number" ref="timeout" v-model:value="job.timeout" placeholder="任务执行超时时间">
      </div>
      <div class="field">
        <label>并行数设置（0 表示不限制）</label>
        <div class="ui icon input">
          <input type="number" ref="parallels" v-model:value="job.parallels" placeholder="任务执行超时时间">
          <i class="large help circle link icon" data-position="top right" data-html="设置在<strong style='color:red'>单个节点</strong>上面同时可执行多少个任务，针对某些任务执行时间很长，但两次任务执行间隔较短时比较有用" data-variation="wide"></i>
        </div>
      </div>
    </div>
    <div class="two fields">
      <div class="field">
        <label>失败重试次数（0 表示不重试）</label>
        <input type="number" ref="retry" v-model:value="job.retry" placeholder="任务失败后重试的次数">
      </div>
      <div class="field">
        <label>失败重试间隔（0 表示立即执行）</label>
        <input type="number" ref="interval" v-model:value="job.interval" placeholder="任务失败后多长时间再次执行">
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
        allowSuffixsTip: '',
        job: {
          id: '',
          kind: 2, // 0 == 普通任务，1 == 单机任务
          name:  '',
          oldGroup: '',
          group: '',
          user: '',
          cmd: '',
          pause: false,
          parallels: 0,
          timeout: 0,
          interval: 0,
          retry: 0,
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

    changeUser: function(val, text){
      this.job.user = val;
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
        .onfailed((resp)=>{vm.$bus.$emit('error', resp)})
        .onend(()=>{vm.loading=false})
        .do();
    },

    newRandomRuleId: function(){
      return 'NEW'+Math.random().toString();
    }
  },

  mounted: function(){
    var vm = this;
    var secCnf = this.$appConfig.security;
    if (secCnf.open) {
      if (secCnf.ext && secCnf.ext.length > 0) {
        this.allowSuffixsTip = '（当前限制只允许添加此类后缀脚本：' + secCnf.ext.join(' ') + '）';
      }
    }

    if (this.$route.path.indexOf('/job/create') === 0) {
      this.action = 'CREATE';
    } else {
      this.action = 'UPDATE';
      this.$rest.GET('job/'+this.$route.params.group+'-'+this.$route.params.id).
        onsucceed(200, (resp)=>{
          vm.job = resp;
          vm.job.oldGroup = resp.group;
          if (vm.job.rules) {
            for (var i in vm.job.rules) {
              if (vm.job.rules[i].id.length == 0) {
                vm.job.rules[i].id = vm.newRandomRuleId();
              }
            }
          }
        }).
        onfailed((msg)=>{vm.error = msg}).
        do();
    }

    this.$rest.GET('job/groups').onsucceed(200, (resp)=>{
      !resp.includes('default') && resp.unshift('default');
      vm.groups = resp;
    }).do();

    $(this.$refs.pause).checkbox({
      onChange: function(){
        vm.job.pause = !vm.job.pause;
      }
    });

    $(this.$refs.kind).find('.checkbox').checkbox({
      onChange: function(){
        vm.job.kind = +$(vm.$refs.kind).find('input[type=radio]:checked').val();
      }
    });

    $(this.$el).find('i.help.icon').popup();
  },

  components: {
    JobEditRule,
    Dropdown
  }
}
</script>