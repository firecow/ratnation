import {expect, test} from "@jest/globals";
import {TraefikTransform} from "../src/stream/traefik-transform.js";
import {pipeline} from "stream/promises";
import {Readable} from "stream";
import {streamToString} from "../src/utils.js";

test("Transforms traefik output to ECS ndjson", async () => {
    const readable = new Readable();
    readable.push("time=\"2023-03-12T15:48:19+01:00\" level=error msg=\"close tcp [::]:2184: use of closed network connection\" entryPointName=tcp");
    readable.push(null);
    const traefikTransform = new TraefikTransform();
    await pipeline(readable, traefikTransform);
    const output = await streamToString(traefikTransform);

    expect(output).toBe(JSON.stringify({
        "@timestamp": "2023-03-12T14:48:19.000Z",
        "log.level": "error",
        "message": "close tcp [::]:2184: use of closed network connection",
        "process.title": "traefik",
    }) + "\n");
});

test("Throws error on unparsable output ", async () => {
    const readable = new Readable();
    readable.push("rubbish");
    readable.push(null);
    const traefikTransform = new TraefikTransform();
    await expect(pipeline(readable, traefikTransform)).rejects.toThrow("TraefikTransform didn't not parse traefik output correctly");
});
