import {TASK_KIND, TASK_STATUS} from "../../Constant";
import {msToTime, parseISOString} from "../../Utils";

export const processData = (blocks) => {
    blocks.map(w => {
        w.task_list.map((t, idx) => {
            t.kind = TASK_KIND[t.kind]
            t.status = TASK_STATUS[t.status]
            t.task_duration = msToTime(t.task_duration / 1e+6)
            t.no = idx + 1
            // t.task_started = parseISOString(t.task_started)
            // t.task_submitted = parseISOString(t.task_submitted)
            // t.task_completed = parseISOString(t.task_completed)
            t.worker = t.worker.split("-").slice(1).join("-")

            if (t.task_transcode != null) {
                t.target = t.task_transcode.target_res + " - " + t.task_transcode.video.file_name.split("-").reverse()[0].split(".")[0]
            } else if (t.task_split != null) {
                if (t.task_split.splited_video != null) {
                    t.target = t.task_split.splited_video.length
                }
            } else if (t.task_merge != null) {
                if (t.task_merge.list_video != null) {
                    t.target = t.task_merge.list_video.length + " - " + t.task_merge.list_video[0].file_name.split("_").reverse()[0].split(".")[0]
                }
            }
        })
    })
    return blocks.reverse()
}