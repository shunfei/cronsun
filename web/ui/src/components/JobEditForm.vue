<template>
  <form class="ui form" v-bind:class="{loading:loading}" v-on:submit.prevent>
    <h3 class="ui header">{{$L(action == 'CREATE' ? 'create job' : 'update job')}}&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
      <em v-if="job.id">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; ID# {{job.id}}</em>
      <div class="ui toggle checkbox" ref="pause">
        <input type="checkbox" class="hidden" v-bind:checked="!job.pause">
        <label v-bind:style="{color: (job.pause?'red':'green')+' !important'}">{{$L(job.pause ? 'pause' : 'open')}}</label>
      </div>&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
      <div class="ui toggle checkbox" ref="fail_notify" v-if="$appConfig.alarm">
        <input type="checkbox" class="hidden" v-bind:checked="job.fail_notify">
        <label>{{$L(job.fail_notify ? 'warning on' : 'warning off')}}</label>
      </div>
    </h3>
    <div class="inline fields" ref="kind">
      <label>{{$L('job type')}}</label>
      <div class="field">
        <div class="ui radio checkbox">
          <input type="radio" v-model="job.kind" name="kind" value="0" tabindex="0" class="hidden"/>
          <label>{{$L('common job')}}</label>
        </div>
      </div>
      <div class="field">
        <div class="ui radio checkbox">
          <input type="radio" v-model="job.kind" name="kind" value="1" tabindex="0" class="hidden"/>
          <label>{{$L('single node single process')}}</label>
        </div>
      </div>
      <div class="field">
        <div class="ui radio checkbox">
          <input type="radio" v-model="job.kind" name="kind" value="2" tabindex="0" class="hidden"/>
          <label>{{$L('group level common')}}
            <i class="help circle link icon" data-position="top right" :data-html="$L('group level common help')" data-variation="wide"></i>
          </label>
        </div>
      </div>
    </div>
    <div class="two fields">
      <div class="field">
        <label>{{$L('job name')}}</label>
        <input type="text" ref="name" v-bind:value="job.name" v-on:input="updateValue($event.target.value)" :placeholder="$L('job name')">
      </div>
      <div class="field">
        <label>{{$L('job group')}}</label>
        <Dropdown :title="$L('select a group')" v-bind:allowAdditions="true" v-bind:items="groups" v-bind:selected="job.group" v-on:change="changeGroup"></Dropdown>
      </div>
    </div>
    <div class="fields">
      <div class="twelve wide field">
        <label>{{$L('script path')}} {{allowSuffixsTip}}</label>
        <input type="text" v-model="job.cmd" :placeholder="$L('script path')">
      </div>
      <div class="four wide field">
        <label>{{$L($appConfig.security.open ? 'user(required)' : 'user(optional)')}}</label>
        <Dropdown v-if="$appConfig.security.open" :title="$L('the user which to execute the command')" v-bind:items="$appConfig.security.users" v-bind:selected="job.user" v-on:change="changeUser"></Dropdown>
        <input v-else type="text" v-model="job.user" :placeholder="$L('the user which to execute the command')">
      </div>
    </div>
    <div class="field" v-if="$appConfig.alarm && job.fail_notify">
      <label>{{$L('warning receiver')}}</label>
      <Dropdown :title="$L('e-mail address')" v-bind:items="alarmReceivers" v-bind:selected="job.to" v-on:change="changeAlarmReceiver" v-bind:multiple="true" v-bind:allowAdditions="true"/>
    </div>
    <div class="two fields">
      <div class="field">
        <label>{{$L('timeout(in seconds, 0 for no limits)')}}</label>
        <input type="number" ref="timeout" v-model.number="job.timeout">
      </div>
      <div class="field" v-show="job.kind === 0">
        <label>{{$L('parallel number in one node(0 for no limits)')}}</label>
        <input type="number" ref="parallels" v-model.number="job.parallels">
      </div>
    </div>
    <div class="two fields">
      <div class="field">
        <label>{{$L('retries(number of retries when failed, 0 means no retry)')}}</label>
        <input type="number" ref="retry" v-model.number="job.retry">
      </div>
      <div class="field">
        <label>{{$L('retry interval(in seconds)')}}</label>
        <input type="number" ref="interval" v-model.number="job.interval">
      </div>
    </div>
    <div class="two fields" v-if="$appConfig.log_expiration_days>0">
      <div class="field">
        <label>{{$L('log expiration(log expired after N days, 0 will use default setting: {n} days)', $appConfig.log_expiration_days)}}</label>
        <input type="number" ref="log_expiration" v-model.number="job.log_expiration">
      </div>
    </div>
    <div class="field">
      <span v-if="!job.rules || job.rules.length == 0"><i class="warning circle icon"></i>{{$L('the job dose not have a timer currently, please click the button below to add a timer')}}</span>
    </div>
    <JobEditRule v-for="(rule, index) in job.rules" :key="rule.id" v-bind:rule="rule" :index="index" v-on:remove="removeRule" v-on:change="changeRule"/>
    <div class="two fields">
      <div class="field">
        <button class="fluid ui button" v-on:click="addNewTimer" type="button"><i class="history icon"></i> {{$L('add timer')}}</button>
      </div>
      <div class="field">
        <button class="fluid blue ui button" type="button" v-on:click="submit"><i class="upload icon"></i> {{$L('save job')}}</button>
      </div>
    </div>
  </form>
</template>

<script>
import JobEditRule from './JobEditRule.vue';
import Dropdown from './basic/Dropdown.vue';
import {split} from '../libraries/functions';

export default {
  name: 'job-edit-form',
  data: function(){
      return {
        action: 'CREATE',
        groups: [],
        alarmReceivers: [],
        loading: false,
        allowSuffixsTip: '',
        job: {
          id: '',
          kind: 0, // 0 == 普通任务，1 == 单机任务
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
          rules: [],
          fail_notify: false,
          log_expiration: 0,
          to: []
        }
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

    changeAlarmReceiver: function(val, text){
      this.job.to = split(val, ',');
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
    var secCnf = vm.$appConfig.security;
    if (secCnf.open) {
      if (secCnf.ext && secCnf.ext.length > 0) {
        vm.allowSuffixsTip = vm.$L('(only [{.suffixs}] files can be allowed)', secCnf.ext.join(' '));
      }
    }

    if (vm.$route.path.indexOf('/job/create') === 0) {
      vm.action = 'CREATE';
    } else {
      vm.action = 'UPDATE';
      vm.$rest.GET('job/'+vm.$route.params.group+'-'+vm.$route.params.id).
        onsucceed(200, (resp)=>{
          vm.job = resp;
          vm.alarmReceivers = resp.to;
          vm.job.oldGroup = resp.group;
          if (vm.job.rules) {
            for (var i in vm.job.rules) {
              if (vm.job.rules[i].id.length == 0) {
                vm.job.rules[i].id = vm.newRandomRuleId();
              }
            }
          }
        }).
        onfailed((msg)=> vm.$bus.$emit('error', data)).
        do();
    }

    vm.$rest.GET('job/groups').onsucceed(200, (resp)=>{
      !resp.includes('default') && resp.unshift('default');
      vm.groups = resp;
    }).do();

    $(vm.$refs.pause).checkbox({
      onChange: function(){
        vm.job.pause = !vm.job.pause;
      }
    });

    $(vm.$refs.fail_notify).checkbox({
      onChange: function(){
        vm.job.fail_notify = !vm.job.fail_notify;
      }
    });

    $(vm.$refs.kind).find('.checkbox').checkbox({
      onChange: function(){
        vm.job.kind = +$(vm.$refs.kind).find('input[type=radio]:checked').val();
      }
    });

    $(vm.$el).find('i.help.icon').popup();
  },

  components: {
    JobEditRule,
    Dropdown
  }
}
</script>
