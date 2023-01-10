export class Config {

    ratholeMap = new Map();

    constructor(argv) {

        for (const ratholeArg of argv["rathole"] ?? []) {
            const ratholeCnf = {};
            const pairs = ratholeArg.split(" ");
            for (const pair of pairs) {
                const key = pair.split("=")[0];
                ratholeCnf[key] = pair.split("=")[1];
            }
            ratholeCnf["bind_port"] = Number(ratholeCnf["bind_port"]);
            this.ratholeMap.set(ratholeCnf["bind_port"], ratholeCnf);
        }

        if (this.ratholeMap.size === 0) {
            console.error("King must have a least one rathole defined in cli options");
            process.exit(1);
        }
    }
}
