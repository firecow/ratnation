import {AssertionError} from "assert";
import isPortReachable from "is-port-reachable";

export async function portsReachable (iter: Iterable<{bind_port: number}>) {
    for (const entry of iter) {
        if (await isPortReachable(entry.bind_port, {host: "0.0.0.0"})) {
            throw new AssertionError({message: `${entry.bind_port} is already in use`});
        }
    }
}
