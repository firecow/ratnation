import {Transform, TransformCallback} from "stream";
import assert from "assert";

export class TraefikTransform extends Transform {

    _transform (chunk: Buffer, encoding: BufferEncoding, callback: TransformCallback) {
        const res = /time="(?<time>.*?)" level=(?<level>.*?) msg="(?<msg>.*?)"/g.exec(`${chunk}`);
        if (res) {
            assert(res.groups != null && res.groups.time);
            callback(null, JSON.stringify({
                "@timestamp": new Date(res.groups.time).toISOString(),
                "log.level": res.groups?.level.toLowerCase(),
                "message": res.groups?.msg,
                "process.title": "traefik",
            }) + "\n");
        } else {
            callback(new Error("TraefikTransform didn't not parse traefik output correctly"));
        }
    }
}