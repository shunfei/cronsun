<style scoped>
.node {
  width: 130px;
  border-radius: 3px;
  padding: 4px 0;
  margin: 3px;
  display: inline-block;
  background: #e8e8e8;
  text-align: center;
}
</style>
<template>
  <div>
    <div class="clearfix">
      <router-link class="ui right floated primary button" to="/node/group"><i class="cubes icon"></i> 管理分组</router-link>
      <div class="ui label" title="正常运行的节点"><i class="green cube icon"></i> {{runningCount}} 正常节点</div>
      <div class="ui label" title="手动下线/维护中的"><i class="cube icon"></i> {{offlineCount}} 离线节点</div>
      <div class="ui label" title="因自身或网络等原因未检测到节点存活"><i class="red cube icon"></i> {{faultCount}} 故障节点</div>
      （总 {{count}} 个节点）
    </div>
  
    <h4 v-if="faultCount > 0" class="ui horizontal divider header"><i class="red cube icon"></i> 故障节点 {{faultCount}}</h4>
    <div v-if="faultCount > 0" v-for="node in faults" class="node">{{node.id}}</div>

    <h4 v-if="offlineCount > 0" class="ui horizontal divider header"><i class="cube icon"></i> 离线节点 {{offlineCount}}</h4>
    <div v-if="offlineCount > 0" v-for="node in offlines" class="node">{{node.id}}</div>

    <h4 v-if="runningCount > 0" class="ui horizontal divider header"><i class="green cube icon"></i> 正常节点 {{runningCount}}</h4>
    <div v-if="runningCount > 0" v-for="node in runnings" class="node">{{node.id}}</div>
  </div>
</template>

<script>
export default {
  name: 'node',
  data: function(){
    return {
      count: 0,
      runningCount: 0,
      offlineCount: 0,
      faultCount: 0,
      runnings: [],
      offlines: [],
      faults: []
    }
  },

  mounted: function(){
    var vm = this;
    this.$rest.GET('nodes').onsucceed(200, (resp)=>{
      resp = [{"id":"192.168.2.16","pid":"8696","alived":true,"connected":true},{"id":"192.168.2.15","pid":"9221","alived":true,"connected":true},{"id":"192.168.2.64","pid":"15831","alived":true,"connected":true},{"id":"192.168.2.77","pid":"18888","alived":true,"connected":true},{"id":"192.168.2.78","pid":"23933","alived":true,"connected":true},{"id":"192.168.2.79","pid":"27693","alived":true,"connected":true},{"id":"192.168.2.80","pid":"5705","alived":true,"connected":true},{"id":"192.168.2.81","pid":"7710","alived":true,"connected":true},{"id":"192.168.2.82","pid":"27368","alived":true,"connected":true},{"id":"192.168.2.83","pid":"7729","alived":true,"connected":true},{"id":"192.168.2.84","pid":"13253","alived":true,"connected":true},{"id":"192.168.2.85","pid":"19747","alived":true,"connected":true},{"id":"192.168.2.86","pid":"25342","alived":true,"connected":true},{"id":"192.168.2.87","pid":"23107","alived":true,"connected":true},{"id":"192.168.2.88","pid":"3232","alived":true,"connected":true},{"id":"192.168.2.59","pid":"4148","alived":true,"connected":true},{"id":"192.168.2.60","pid":"18070","alived":true,"connected":true},{"id":"192.168.2.61","pid":"5888","alived":true,"connected":true},{"id":"192.168.2.62","pid":"8967","alived":true,"connected":true},{"id":"192.168.2.63","pid":"30765","alived":true,"connected":true},{"id":"192.168.2.65","pid":"6697","alived":true,"connected":true},{"id":"192.168.2.66","pid":"23452","alived":true,"connected":true},{"id":"192.168.2.67","pid":"23435","alived":true,"connected":true},{"id":"192.168.2.68","pid":"31018","alived":true,"connected":true},{"id":"192.168.2.69","pid":"30195","alived":true,"connected":true},{"id":"192.168.2.70","pid":"1204","alived":true,"connected":true},{"id":"192.168.2.47","pid":"25954","alived":true,"connected":true},{"id":"192.168.2.49","pid":"21290","alived":true,"connected":true},{"id":"192.168.2.48","pid":"4506","alived":true,"connected":true},{"id":"192.168.2.50","pid":"2869","alived":true,"connected":true},{"id":"192.168.2.51","pid":"24557","alived":true,"connected":true},{"id":"192.168.2.52","pid":"32352","alived":true,"connected":true},{"id":"192.168.2.53","pid":"21321","alived":true,"connected":true},{"id":"192.168.2.54","pid":"13271","alived":true,"connected":true},{"id":"192.168.2.55","pid":"10954","alived":true,"connected":true},{"id":"192.168.2.56","pid":"14497","alived":true,"connected":true},{"id":"192.168.2.57","pid":"1068","alived":true,"connected":true},{"id":"192.168.2.58","pid":"16134","alived":true,"connected":true},{"id":"192.168.2.32","pid":"10599","alived":true,"connected":true},{"id":"192.168.2.33","pid":"7409","alived":true,"connected":true},{"id":"192.168.2.34","pid":"18769","alived":true,"connected":true},{"id":"192.168.2.35","pid":"27350","alived":true,"connected":true},{"id":"192.168.2.36","pid":"11909","alived":true,"connected":true},{"id":"192.168.2.37","pid":"8722","alived":true,"connected":true},{"id":"192.168.2.38","pid":"23674","alived":true,"connected":true},{"id":"192.168.2.39","pid":"29680","alived":true,"connected":true},{"id":"192.168.2.40","pid":"9223","alived":true,"connected":true},{"id":"192.168.2.41","pid":"1062","alived":true,"connected":true},{"id":"192.168.2.17","pid":"11287","alived":true,"connected":true},{"id":"192.168.2.18","pid":"6212","alived":true,"connected":true},{"id":"192.168.2.19","pid":"1413","alived":true,"connected":true},{"id":"192.168.2.20","pid":"13118","alived":true,"connected":true},{"id":"192.168.2.21","pid":"5829","alived":true,"connected":true},{"id":"192.168.2.22","pid":"23197","alived":true,"connected":true},{"id":"192.168.2.23","pid":"21841","alived":true,"connected":true},{"id":"192.168.2.24","pid":"16373","alived":true,"connected":true},{"id":"192.168.2.25","pid":"17516","alived":true,"connected":true},{"id":"192.168.2.26","pid":"29610","alived":false,"connected":false},{"id":"192.168.2.2","pid":"14070","alived":true,"connected":true},{"id":"192.168.2.3","pid":"15851","alived":true,"connected":true},{"id":"192.168.2.4","pid":"11543","alived":true,"connected":true},{"id":"192.168.2.5","pid":"21092","alived":true,"connected":true},{"id":"192.168.2.6","pid":"32198","alived":true,"connected":true},{"id":"192.168.2.7","pid":"5441","alived":true,"connected":true},{"id":"192.168.2.8","pid":"18684","alived":true,"connected":true},{"id":"192.168.2.9","pid":"10570","alived":true,"connected":true},{"id":"192.168.2.10","pid":"19416","alived":true,"connected":true},{"id":"192.168.2.11","pid":"21754","alived":true,"connected":true},{"id":"192.168.2.12","pid":"31135","alived":true,"connected":true},{"id":"192.168.2.13","pid":"1812","alived":true,"connected":true},{"id":"192.168.2.14","pid":"10007","alived":true,"connected":true},{"id":"192.168.2.92","pid":"4207","alived":true,"connected":true},{"id":"192.168.2.93","pid":"25329","alived":true,"connected":true},{"id":"192.168.2.94","pid":"26739","alived":true,"connected":true},{"id":"192.168.2.95","pid":"2648","alived":true,"connected":true},{"id":"192.168.2.96","pid":"15946","alived":true,"connected":true},{"id":"192.168.2.97","pid":"15460","alived":true,"connected":true},{"id":"192.168.2.98","pid":"25245","alived":true,"connected":true},{"id":"192.168.2.99","pid":"27027","alived":true,"connected":true},{"id":"192.168.2.100","pid":"21402","alived":true,"connected":true},{"id":"192.168.2.101","pid":"9065","alived":true,"connected":true},{"id":"192.168.2.102","pid":"1314","alived":true,"connected":true},{"id":"192.168.2.103","pid":"14352","alived":true,"connected":true},{"id":"192.168.2.104","pid":"23860","alived":true,"connected":true},{"id":"192.168.2.105","pid":"10934","alived":true,"connected":true},{"id":"192.168.2.106","pid":"18130","alived":true,"connected":true},{"id":"192.168.2.107","pid":"19025","alived":true,"connected":true},{"id":"192.168.2.108","pid":"17315","alived":true,"connected":true},{"id":"192.168.2.109","pid":"7386","alived":true,"connected":true},{"id":"192.168.2.110","pid":"22908","alived":true,"connected":true},{"id":"192.168.2.111","pid":"15905","alived":true,"connected":true},{"id":"192.168.2.112","pid":"29847","alived":true,"connected":true},{"id":"192.168.0.74","pid":"12161","alived":true,"connected":true},{"id":"192.168.0.75","pid":"11990","alived":true,"connected":true},{"id":"192.168.0.76","pid":"20720","alived":true,"connected":true},{"id":"192.168.0.77","pid":"8076","alived":true,"connected":true},{"id":"192.168.0.78","pid":"11214","alived":true,"connected":true},{"id":"192.168.0.79","pid":"18587","alived":true,"connected":true},{"id":"192.168.0.80","pid":"23014","alived":true,"connected":true},{"id":"192.168.0.81","pid":"29506","alived":true,"connected":true},{"id":"192.168.0.83","pid":"15779","alived":true,"connected":true},{"id":"192.168.0.62","pid":"14641","alived":true,"connected":true},{"id":"192.168.0.63","pid":"14652","alived":true,"connected":true},{"id":"192.168.0.64","pid":"27370","alived":true,"connected":true},{"id":"192.168.0.65","pid":"10387","alived":true,"connected":true},{"id":"192.168.0.84","pid":"27207","alived":true,"connected":true},{"id":"192.168.0.66","pid":"32302","alived":true,"connected":true},{"id":"192.168.0.67","pid":"31941","alived":true,"connected":true},{"id":"192.168.0.68","pid":"8381","alived":true,"connected":true},{"id":"192.168.0.69","pid":"25909","alived":true,"connected":true},{"id":"192.168.0.70","pid":"13360","alived":true,"connected":true},{"id":"192.168.0.71","pid":"14979","alived":true,"connected":true},{"id":"192.168.0.72","pid":"796","alived":true,"connected":true},{"id":"192.168.0.73","pid":"15599","alived":true,"connected":true},{"id":"192.168.0.47","pid":"20656","alived":true,"connected":true},{"id":"192.168.0.48","pid":"25365","alived":true,"connected":true},{"id":"192.168.0.49","pid":"6819","alived":true,"connected":true},{"id":"192.168.0.17","pid":"16979","alived":true,"connected":true},{"id":"192.168.0.18","pid":"11104","alived":true,"connected":true},{"id":"192.168.0.20","pid":"24509","alived":true,"connected":true},{"id":"192.168.0.200","pid":"24846","alived":true,"connected":true},{"id":"192.168.0.19","pid":"22238","alived":true,"connected":true},{"id":"192.168.0.23","pid":"13002","alived":true,"connected":true},{"id":"192.168.0.22","pid":"28269","alived":true,"connected":true},{"id":"192.168.0.25","pid":"32589","alived":true,"connected":true},{"id":"192.168.0.24","pid":"12671","alived":true,"connected":true},{"id":"192.168.0.26","pid":"24109","alived":true,"connected":true},{"id":"192.168.0.28","pid":"10255","alived":true,"connected":true},{"id":"192.168.0.27","pid":"2534","alived":true,"connected":true},{"id":"192.168.0.32","pid":"16626","alived":true,"connected":true},{"id":"192.168.0.33","pid":"20208","alived":true,"connected":true},{"id":"192.168.0.35","pid":"32108","alived":true,"connected":true},{"id":"192.168.0.36","pid":"3304","alived":true,"connected":true},{"id":"192.168.0.34","pid":"19218","alived":true,"connected":true},{"id":"192.168.0.37","pid":"29217","alived":true,"connected":true},{"id":"192.168.0.38","pid":"6163","alived":true,"connected":true},{"id":"192.168.0.39","pid":"11260","alived":true,"connected":true},{"id":"192.168.0.40","pid":"23025","alived":true,"connected":true},{"id":"192.168.0.41","pid":"23994","alived":true,"connected":true},{"id":"192.168.0.43","pid":"20967","alived":true,"connected":true},{"id":"192.168.0.42","pid":"12863","alived":true,"connected":true},{"id":"192.168.0.44","pid":"14474","alived":true,"connected":true},{"id":"192.168.0.2","pid":"7016","alived":true,"connected":true},{"id":"192.168.0.45","pid":"23905","alived":true,"connected":true},{"id":"192.168.0.3","pid":"2416","alived":true,"connected":true},{"id":"192.168.0.4","pid":"11511","alived":true,"connected":true},{"id":"192.168.0.5","pid":"5236","alived":true,"connected":true},{"id":"192.168.0.6","pid":"12590","alived":true,"connected":true},{"id":"192.168.0.8","pid":"32300","alived":true,"connected":true},{"id":"192.168.0.7","pid":"1917","alived":true,"connected":true},{"id":"192.168.0.9","pid":"1744","alived":true,"connected":true},{"id":"192.168.0.10","pid":"26385","alived":true,"connected":true},{"id":"192.168.0.11","pid":"3720","alived":true,"connected":true},{"id":"192.168.0.12","pid":"2052","alived":true,"connected":true},{"id":"192.168.0.13","pid":"23372","alived":true,"connected":true},{"id":"192.168.0.14","pid":"8418","alived":true,"connected":true},{"id":"192.168.0.15","pid":"30412","alived":true,"connected":true},{"id":"192.168.0.16","pid":"3280","alived":true,"connected":true}]
      for (var i in resp) {
        var n = resp[i];
        if (n.alived && n.connected) {
          vm.runnings.push(n);
        } else if (n.alived && !n.connected) {
          vm.faults.push(n);
        } else {
          vm.offlines.push(n);
        }
      }
      vm.runningCount = vm.runnings.length;
      vm.offlineCount = vm.offlines.length;
      vm.faultCount = vm.faults.length;
      vm.count = resp.length || 0;
    }).do();
  }
}
</script>