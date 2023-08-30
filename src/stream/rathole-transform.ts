import {Transform, TransformCallback} from "stream";
import assert from "assert";

export class RatholeTransform extends Transform {

    _transform (chunk: Buffer, encoding: BufferEncoding, callback: TransformCallback) {
        const res = /(?<time>.*?) {2}(?<level>.*?) (?<msg>.*)/g.exec(`${chunk}`);
        if (res) {
            assert(res.groups?.time != null);
            const timeWithZone = new Date(res.groups.time);
            timeWithZone.setMinutes(timeWithZone.getMinutes() - timeWithZone.getTimezoneOffset());
            callback(null, JSON.stringify({
                "@timestamp": new Date(timeWithZone).toISOString(),
                "log.level": res.groups?.level.toLowerCase(),
                "message": res.groups?.msg,
                "process.title": "rathole",
            }) + "\n");
        } else {
            callback(new Error("RatholeTransform didn't not parse rathole output correctly"));
        }
    }
}