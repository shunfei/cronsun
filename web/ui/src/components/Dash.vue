<style scope>
#numbers {padding-bottom: 30px;}
#numbers .card .header {text-align: center;}
#numbers .card .number {font-size: 2em;}
#numbers .card .title {font-size: 1em;}

#charts>div {
    width: 300px; 
    display: inline-block;
    margin: 0 1em;
    border: none;
    box-shadow: none;
}
#charts .header {text-align: center;}
</style>

<template>
  <div>
    <div id="numbers" class="ui three column grid">
      <div class="row">
        <div class="column">
          <div class="ui fluid card">
            <div class="content">
              <div class="header number">{{totalJobs}}</div>
              <div class="header title">任务总数</div>
            </div>
          </div>
        </div>
        <div class="column">
          <div class="ui fluid card">
            <div class="content">
              <div class="header number">{{totalExecuted}}</div>
              <div class="header title">执行任务总次数</div>
            </div>
          </div>
        </div>
        <div class="column">
          <div class="ui fluid card">
            <div class="content">
              <div class="header number">{{todayExecuted}}</div>
              <div class="header title">今日执行任务次数</div>
            </div>
          </div>
        </div>
      </div>
      <div class="row">
        <div class="column">
          <div class="ui fluid card">
            <div class="content">
              <div class="header number">{{totalNodes}}</div>
              <div class="header title">节点总数</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div id="charts">
      <div class="ui card">
        <div class="content">
          <h4 class="header"><router-link to="node">当前节点状态</router-link></h4>
          <div class="description">
            <canvas ref="node"></canvas>
          </div>
        </div>
      </div>

      <div class="ui card">
        <div class="content">
          <h4 class="header"><router-link :to="'log?begin='+today+'&end='+today">今日任务概况</router-link></h4>
          <div class="description">
            <canvas ref="job"></canvas>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Chart from 'charts';
import {formatNumber} from '../libraries/functions';

export default {
  name: 'dash',
  data(){
    return {
      totalJobs: 0,
      totalExecuted: 0,
      todayExecuted: 0,
      totalNodes: 0,

      today: ''
    }
  },

  mounted(){
    // var formatNumber = function(i, len){
    //   var n = i == 0 ? 1 : Math.ceil(Math.log10(i+1));
    //   if (n >= len) return i.toString();
    //   return '0'.repeat(len-n) + i.toString(); 
    // }
    var d = new Date()
    this.today = d.getFullYear().toString() + '-' + formatNumber(d.getMonth()+1, 2) + '-' + d.getDate();

    var vm = this;
    var renderJobInfo = function(resp){
      vm.totalJobs = resp.totalJobs;
      vm.totalExecuted = resp.jobExecuted ? resp.jobExecuted.total : 0;
      vm.todayExecuted = resp.jobExecutedDaily ? resp.jobExecutedDaily.total : 0;
      var successed = resp.jobExecuted ? resp.jobExecuted.successed : 0;
      var failed = resp.jobExecuted ? resp.jobExecuted.failed : 0;
      new Chart($(vm.$refs.job), {
        type: 'pie',
        data: {
          labels: ["成功", "失败"],
          datasets: [{
          data: [successed, failed],
            backgroundColor: ["#21BA45", "#DB2828"],
            hoverBackgroundColor: ["#39DE60", "#D64848"]
          }]
        }
      });
    }

    var renderNodeInfo = function(resp){
      vm.totalNodes = resp ? resp.length : 0;
      var online = 0;
      var offline = 0;
      var damaged = 0;
      for (var i in resp) {
        if (resp[i].alived && resp[i].connected) {
          online++;
        } else if (resp[i].alived && !resp[i].connected) {
          damaged++;
        } else if(!resp[i].alived) {
          offline++;
        }
      }
      
      new Chart($(vm.$refs.node), {
        type: 'pie',
        data: {
          labels: ["在线", "离线", "故障"],
          datasets: [{
          data: [online, offline, damaged],
            backgroundColor: ["#21BA45", "#333", "#DB2828"],
            hoverBackgroundColor: ["#39DE60", "#555", "#D64848"]
          }]
        }
      });
    }

    this.$rest.GET('/info/overview').onsucceed(200, renderJobInfo).do();
    this.$rest.GET('nodes').onsucceed(200, renderNodeInfo).do();
  }
}
</script>