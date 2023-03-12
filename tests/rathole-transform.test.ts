import {expect, test} from "@jest/globals";
import {pipeline} from "stream/promises";
import {Readable} from "stream";
import {RatholeTransform} from "../src/rathole-transform.mjs";
import {streamToString} from "../src/utils.mjs";

test("Transforms rathole output to ECS ndjson", async () => {
    const readable = new Readable();
    readable.push("Mar 12 15:20:47.730  INFO handle{service=c94b1d98-70f8-459b-84a0-044dc8fcd0af}: rathole::client: Starting e7240b0a882049e4d5a21f770f069887b31e4573940f4a1cc370f87ae1425975");
    readable.push(null);
    const ratholeTransform = new RatholeTransform();
    await pipeline(readable, ratholeTransform);
    const output = await streamToString(ratholeTransform);

    expect(output).toBe(JSON.stringify({
        "@timestamp": "2001-03-12T14:20:47.730Z",
        "log.level": "info",
        "message": "handle{service=c94b1d98-70f8-459b-84a0-044dc8fcd0af}: rathole::client: Starting e7240b0a882049e4d5a21f770f069887b31e4573940f4a1cc370f87ae1425975",
        "process.title": "rathole",
    }) + "\n");
});

test("Throws error on unparsable output ", async () => {
    const readable = new Readable();
    readable.push("rubbish");
    readable.push(null);
    const traefikTransform = new RatholeTransform();
    await expect(pipeline(readable, traefikTransform)).rejects.toThrow("RatholeTransform didn't not parse rathole output correctly");
});