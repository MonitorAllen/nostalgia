<script setup lang="ts">
import {useUserStore} from "@/store/module/user";

// @ts-ignore
import CalHeatmap from "cal-heatmap";
// @ts-ignore
import LegendLite from 'cal-heatmap/plugins/LegendLite';
// @ts-ignore
import Tooltip from 'cal-heatmap/plugins/Tooltip';
// @ts-ignore
import CalendarLabel from 'cal-heatmap/plugins/CalendarLabel';

import 'cal-heatmap/cal-heatmap.css'

const userStore = useUserStore()

const heatmap: CalHeatmap = new CalHeatmap()

userStore.contributions()
    .then((res: any) => {
      const contributions = res.data.contributions.reverse()
      heatmap.paint({
            data: {
              source: contributions,
              type: 'json',
              x: 'date',
              y: (d: any) => +d['intensity'],
            },
            date: {
              start: new Date(contributions[0].date),
              max: new Date(Date.now())
            },
            range: 4,
            domain: {
              type: 'month',
              gutter: 4,
              label: {
                text: 'MMM',
                position: 'top',
                textAlign: 'start',
                offset: {y: 4}
              }
            },
            subDomain: {
              type: 'ghDay',
            },
            scale: {
              color: {
                type: 'threshold',
                range: ["#ebedf0", "#9be9a8", "#40c463", "#30a14e", "#216e39"],
                domain: [1, 3, 5, 10],
              }
            },
          },
          [
            [
              Tooltip,
              {
                text: function (date: any, value: number, dayjsDate: any) {
                  return (value ? '在 ' + dayjsDate.format('YYYY-MM-DD') + ' 有 ' + value + ' 次活动' : dayjsDate.format('YYYY-MM-DD') + ' 这一天很懒')
                }
              }
            ],
            [
              LegendLite,
              {
                itemSelector: '#ex-ghDay-legend',
                radius: 0,
                width: 11,
                height: 11,
                gutter: 0,
              },
            ],
            [
              CalendarLabel,
              {
                width: 30,
                textAlign: 'start',
                // @ts-ignore
                text: () => dayjs.weekdaysShort().map((d: any, i: any) => (i % 2 == 0 ? '' : d)),
                padding: [25, 0, 0, 0],
              },
            ],
          ],)
    })

</script>

<template>
  <div class="flex flex-column justify-content-center align-items-center">
    <div id="cal-heatmap" class="flex"></div>
    <div class="w-full flex justify-content-end align-items-center flex-wrap mt-2 text-xs text-color-secondary">
      <span>Less</span>
      <div id="ex-ghDay-legend" class="mx-2"></div>
      <span>More</span>
    </div>
  </div>
</template>

<style scoped>

</style>