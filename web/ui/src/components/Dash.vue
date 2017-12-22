<style scope>
#numbers {padding-bottom: 30px;}
#numbers .card .header {text-align: center;}
#numbers .card .number {font-size: 2em;}
#numbers .card .title {font-size: 1em;}
</style>

<template>
  <div>
    <div id="numbers" class="ui three column grid">
      <div class="row">
        <div class="column">
          <div class="ui fluid card">
            <div class="content">
              <div class="header number">
                <span style="color:green;">{{totalOnlineNodes}}</span>/
                <span>{{totalOfflineNodes}}</span>/
                <span style="color:red;">{{totalDamagedNodes}}</span>
              </div>
              <div class="header title">{{$L('total number of nodes')}}</div>
            </div>
          </div>
        </div>
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
      </div>
    </div>

    <div id="charts">
      <canvas ref="daily" height="80"></canvas>
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
      totalOnlineNodes: 0,
      totalOfflineNodes: 0,
      totalDamagedNodes: 0,

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
      var data = {
        labels: [],
        datasets: [{
          label: vm.$L('successed'),
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgb(75, 192, 192)',
          fill: false,
          yAxisID: 'yAxisSuccessed',
          data: []
        }, {
          label: vm.$L('failed'),
            borderColor: 'rgb(255, 99, 132)',
            backgroundColor: 'rgb(255, 99, 132)',
            fill: false,
            yAxisID: 'yAxisFailed',
            data: []
        }]
      };

      for (var i in resp.jobExecutedDaily) {
        var info = resp.jobExecutedDaily[i];
        data.labels.push(info.date),
        data.datasets[0].data.push(info.successed);
        data.datasets[1].data.push(info.failed);
      }

      var ctx = vm.$refs.daily.getContext('2d');
      var chart = Chart.Line(ctx, {
        data: data,
        options: {
          responsive: true,
          hoverMode: 'index',
          stacked: false,
          title:{
            display: true,
            text: vm.$L('job executed in past 7 days')
          },
          scales: {
            yAxes: [{
              type: 'linear',
              display: true,
              position: 'left',
              id: 'yAxisSuccessed'
            }, {
              type: 'linear',
              display: true,
              position: 'right',
              id: 'yAxisFailed',
              gridLines: {
                drawOnChartArea: false
              }
            }],
          }
        }
      });
      chart.update();
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

      vm.totalOnlineNodes = online;
      vm.totalOfflineNodes = offline;
      vm.totalDamagedNodes = damaged;
    }

    this.$rest.GET('/info/overview').onsucceed(200, renderJobInfo).do();
    this.$rest.GET('nodes').onsucceed(200, renderNodeInfo).do();
  }
}
</script>
