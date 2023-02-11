import {execa, ExecaChildProcess} from "execa";
import os from "os";
import waitFor from "p-wait-for";

const hostname = os.hostname();
const subprocesses: ExecaChildProcess[] = [];

async function killSubprocesses () {
    const proms = [];
    for (const p of subprocesses) {
        proms.push(waitFor(() => p.exitCode != null));
    }
    await Promise.all(proms);
}

function startSubprocess (file: string, args: string[]) {
    const p = execa(file, args);
    p.stdout?.pipe(process.stdout);
    p.stderr?.pipe(process.stdout);
    subprocesses.push(p);
}

startSubprocess("docker", ["run", "--rm", "--name=ratnation-echo-server", "-p=3000:8080", "jmalloc/echo-server"]);

startSubprocess("nodemon", ["src/index.mjs", "council"]);

startSubprocess("nodemon", ["src/index.mjs", "king", `--host=${hostname}`, "--rathole=\"bind_port=2334 ports=5000-5001\""]);

startSubprocess("nodemon", ["src/index.mjs", "ling",
    "--ling-id=\"0a976e7a-87c5-4549-9431-e4881c740cec\"",
    "--rathole=\"name=alpha local_addr=localhost:3000\"",
    "--proxy=\"name=alpha bind_port=2183\"",
]);

startSubprocess("nodemon", ["src/index.mjs", "ling",
    "--ling-id=\"0573442b-5491-444e-9c63-c2907079ff5f\"",
    "--rathole=\"name=alpha local_addr=localhost:3000\"",
    "--proxy=\"name=alpha bind_port=2184\"",
]);

process.on("SIGINT", async () => killSubprocesses());
