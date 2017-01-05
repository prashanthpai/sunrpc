struct intpair {
	int a;
	int b;
};

program ARITH_PROG {
	version ARITH_VERS {
		int ADD(intpair) = 1;
		int MULTIPLY(intpair) = 2;
	} = 1;
} = 12345;
