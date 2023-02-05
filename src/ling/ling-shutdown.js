import wait from "wait-promise";

export function initLingShutdownHandlers({context, stateHandler, syncer, traefikManager, ratholeManager}) {
    const listener = async(signal) => {
        console.log("msg=\"ling shutdown sequence initiated\" service.type=ratling");
        context.shuttingDown = true;
        stateHandler.stop();
        await syncer.stop();
        await wait.sleep(1000);
        ratholeManager.killProcesses(signal);
        traefikManager.killProcesses(signal);
    };
    process.on("SIGINT", listener);
    process.on("SIGTERM", listener);
}
