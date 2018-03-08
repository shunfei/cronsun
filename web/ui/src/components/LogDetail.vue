<style scoped>
.title {
  display: inline-block;
  width: 80px;
}
</style>
<template>
  <div v-if="error != ''" class="ui negative message">
    <div class="header"><i class="attention icon"></i> {{error}}</div>
  </div>
  <div v-else>
    <div class="ui segments">
      <div class="ui segment">
        <p>
          <span class="title">{{$L('name')}}</span>
          <router-link class="item" :to="'/job/edit/'+log.jobGroup+'/'+log.jobId">{{log.name}}</router-link></p>
      </div>
      <div class="ui segment">
        <p>
          <span class="title">{{$L('node')}}</span> {{node.hostname}} [{{node.ip}}]
        </p>
      </div>
      <div class="ui segment">
        <p>
           <span class="title">{{$L('user')}}</span>
           <i class="attention warning icon" v-if="log.user == 'root' || log.user == ''"></i> {{log.user}}
        </p>
      </div>
      <div class="ui segment">
        <p>
          <span class="title">{{$L('spend time')}}</span>
          {{log.beginTime}} - {{log.endTime}}
        </p>
      </div>
      <div class="ui segment">
        <p>
          <span class="title">{{$L('result')}}</span>
          <span v-if="log.success"><i class="checkmark green icon"></i></span>
          <span v-else><i class="remove red icon"></i></span>
        </p>
      </div>
    </div>
    <h4 class="ui header">{{$L('command')}}</h4>
    <pre class="ui grey inverted segment">{{log.command}}</pre>
    <h4 class="ui header">{{$L('output')}}</h4>
    <pre class="ui inverted segment">{{printResult}}</pre>
  </div>
</template>

<script>
export default {
  name: 'log-detail',
  data: function(){
      return {
        log: {
          id: '',
          jobId: '',
          jobGroup: '',
          name:  '',
          node:  '',
          user:  '',
          command: '',
          output: '',
          exitCode: 0,
          beginTime: new Date(),
          endTime: new Date()
        },
        node: {},
        error: ''
      }
  },

  computed: {
    printResult(){
      return this.log.output ? this.log.output.replace("\r\n", "^M\r\n") : '';
    }
  },

  mounted: function(){
    var vm = this;
    this.$rest.GET('log/'+this.$route.params.id).
        onsucceed(200, (resp)=>{
          vm.log = resp;
          vm.node = vm.$store.getters.getNodeByID(resp.node)
        }).
        onfailed((data, xhr) => {
          if (xhr.status === 404) {
            vm.error = vm.$L('log has been deleted')
          } else {
            vm.error = data
          }
        }).
        do();
  }
}
</script>
