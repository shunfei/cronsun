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
              <div class="header title">{{$L('total number of jobs')}}</div>
            </div>
          </div>
        </div>
        <div class="column">
          <div class="ui fluid card">
            <div class="content">
              <div class="header number">{{totalExecuted}}</div>
              <div class="header title">{{$L('total number of executeds')}}</div>
            </div>
          </div>
        </div>
        <div class="column">
          <div class="ui fluid card">
            <div class="content">
              <div class="header number">{{todayExecuted}}</div>
              <div class="header title">{{$L('total number of executeds(today)')}}</div>
            </div>
          </div>
        </div>
      </div>
      <div class="row">
        <div class="column">
          <div class="ui fluid card">
            <div class="content">
              <div class="header number">{{totalNodes}}</div>
              <div class="header title">{{$L('total number of nodes')}}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div id="charts">
      <div class="ui card">
        <div class="content">
          <h4 class="header"><router-link to="node">{{$L('node stat')}}</router-link></h4>
          <div class="description">
            <canvas ref="node"></canvas>
          </div>
        </div>
      </div>

      <div class="ui card">
        <div class="content">
          <h4 class="header"><router-link :to="'log?begin='+today+'&end='+today">{{$L('executed stat(today)')}}</router-link></h4>
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
    var d = new Date()
    this.today = d.getFullYear().toString() + '-' + formatNumber(d.getMonth()+1, 2) + '-' + d.getDate();

    var vm = this;
    var renderJobInfo = function(resp){
      vm.totalJobs = resp.totalJobs;
      vm.totalExecuted = resp.jobExecuted ? resp.jobExecuted.total : 0;
      vm.todayExecuted = resp.jobExecutedDaily ? resp.jobExecutedDaily.total : 0;
      var dailySuccessed = resp.jobExecutedDaily ? resp.jobExecutedDaily.successed : 0;
      var dailytotal = resp.jobExecutedDaily ? resp.jobExecutedDaily.total : 0;
      new Chart($(vm.$refs.job), {
        type: 'pie',
        data: {
          labels: [vm.$L("{n} successed", dailySuccessed), vm.$L("{n} failed", dailytotal-dailySuccessed)],
          datasets: [{
            data: [dailySuccessed, dailytotal - dailySuccessed],
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
          labels: [vm.$L("{n} online", online), vm.$L("{n} offline", offline), vm.$L("{n} damaged", damaged)],
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