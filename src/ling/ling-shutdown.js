import wait from "wait-promise";

export function initShutdownHandlers({context, stateHandler, syncer, traefikManager, ratholeManager}) {
    const listener = async(signal) => {
        console.log("msg=\"ling shutdown sequence initiated\" service.type=ratling");
        context.shuttingDown = true;
        await syncer.stop();
        await wait.sleep(500);
        stateHandler.stop();
        ratholeManager.killProcesses(signal);
        traefikManager.killProcesses(signal);
    };
    process.on("SIGINT", listener);
    process.on("SIGTERM", listener);
}
