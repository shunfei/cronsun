<template>
  <div v-if="error != ''" class="ui negative message">
    <div class="header"><i class="attention icon"></i> {{error}}</div>
  </div>
  <div v-else>
    <div class="ui segments">
      <div class="ui segment">
        <p>任务：<router-link class="item" :to="'/job/edit/'+log.jobGroup+'/'+log.jobId">{{log.name}}</router-link></p>
      </div>
      <div class="ui segment">
        <p>节点：{{log.node}}</p>
      </div>
      <div class="ui segment">
        <p>时间：{{log.beginTime}} 到 {{log.endTime}}</p>
      </div>
      <div class="ui segment">
        <p>结果：{{log.success ? '成功' : '失败'}}</p>
      </div>
    </div>
    <h4 class="ui header">执行的命令</h4>
    <pre class="ui grey inverted segment">{{log.command}}</pre>
    <h4 class="ui header">输出</h4>
    <pre class="ui inverted segment">{{log.output}}</pre>
  </div>
</template>

<script>
export default {
  name: 'log-detail',
  data: function(){
      return {
        log: {
          id: 'sdfas',
          jobId: 'wewe',
          jobGroup: 'test',
          name:  'run run run',
          node:  '192.168.1.2',
          command: 'echo hello;',
          output: 'hello',
          exitCode: 0,
          beginTime: new Date(),
          endTime: new Date()
        },
        error: ''
      }
  },

  mounted: function(){
    var vm = this;
    this.$rest.GET('log/'+this.$route.params.id).
        onsucceed(200, (resp)=>{vm.log = resp}).
        onfailed((data)=>{vm.error = data.error}).
        do();
  }
}
</script>
