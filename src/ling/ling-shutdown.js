import wait from "wait-promise";

export function initShutdownHandlers({context, stateHandler, configSyncer, traefikManager, ratholeManager}) {
    const listener = async(signal) => {
        console.log("msg=\"ling shutdown sequence initiated\" service.type=ratling");
        context.shuttingDown = true;
        await wait.sleep(500);
        stateHandler.stop();
        configSyncer.stop();
        ratholeManager.killProcesses(signal);
        traefikManager.killProcesses(signal);
        process.exit(0);
    };
    process.on("SIGINT", listener);
    process.on("SIGTERM", listener);
}
