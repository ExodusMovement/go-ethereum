{
	// count tracks the number of EVM instructions executed.
	events: [],

	// step is invoked for every opcode that the VM executes.
	step: function(log, db) {
    // Capture any errors immediately
		var error = log.getError();
		if (error !== undefined) {
			this.fault(log, db);
			return;
		}

        var op = log.op.toString();
        if (!op.startsWith('LOG')) return;

        var numOfTopics = parseInt(op[3], 10);
        if (numOfTopics < 0 && numOfTopics > 4) return;
        var topics = [];
        for (var i = 1; i <= numOfTopics; i++) {
          topics.push('0x' + log.stack.peek(1 + i).toString(16));
        }

        var inputStart = log.stack.peek(0).valueOf();
        var inputLen = inputStart + log.stack.peek(1).valueOf();
        var data = toHex(log.memory.slice(inputStart, inputLen));
        this.events.push({
          topics,
          data,
          address: toHex(log.contract.getAddress())
        });
  },

	// fault is invoked when the actual execution of an opcode fails.
	fault: function(log, db) {
	    this.events = [{
	        error: log.getError()
        }]

        return this.result();
    },

	// result is invoked when all the opcodes have been iterated over and returns
	// the final result of the tracing.
	result: function(ctx, db) { return this.events }
}