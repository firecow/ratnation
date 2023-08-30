import {AssertionError} from "assert";
import isPortReachable from "is-port-reachable";
import {Readable} from "stream";

export async function portsReachable (iter: Iterable<{bind_port: number}>) {
    for (const entry of iter) {
        if (await isPortReachable(entry.bind_port, {host: "0.0.0.0"})) {
            throw new AssertionError({message: `${entry.bind_port} is already in use`});
        }
    }
}

export async function streamToString (stream: Readable): Promise<string> {
    const chunks: Buffer[] = [];
    return new Promise((resolve, reject) => {
        stream.on("data", (chunk: Buffer) => chunks.push(Buffer.from(chunk)));
        stream.on("error", (err: Error) => reject(err));
        stream.on("end", () => resolve(Buffer.concat(chunks).toString("utf8")));
    });
}

export async function to<T>(promise: Promise<T>) {
    return promise
        .then((data) => [null, data])
        .catch((err) => {
            return [err, undefined];
        });
}
