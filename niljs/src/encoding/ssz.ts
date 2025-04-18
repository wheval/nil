import {
  BooleanType,
  ByteListType,
  ByteVectorType,
  ContainerType,
  UintBigintType,
  UintNumberType,
} from "@chainsafe/ssz";

/**
 * The basic types used in the library.
 *
 */
const basicTypes = {
  Uint8: new UintNumberType(1),
  Uint32: new UintNumberType(4),
  Uint64: new UintNumberType(8),
  UintBn256: new UintBigintType(32),
  Bool: new BooleanType(),
};

/**
 * The const representing a byte vector with 20 elements.
 *
 */
const Bytes20 = new ByteVectorType(20);

/**
 * The SSZ schema for a transaction object.
 */
const SszTransactionSchema = new ContainerType({
  deploy: basicTypes.Bool,
  feeCredit: basicTypes.UintBn256,
  maxPriorityFeePerGas: basicTypes.UintBn256,
  maxFeePerGas: basicTypes.UintBn256,
  to: Bytes20,
  chainId: basicTypes.Uint64,
  seqno: basicTypes.Uint64,
  data: new ByteListType(24576),
});

/**
 * SSZ schema for a signed transaction object. Includes auth data in addition to all other transaction fields.
 */
const SszSignedTransactionSchema = new ContainerType({
  ...SszTransactionSchema.fields,
  authData: new ByteListType(256),
});

export { SszTransactionSchema, SszSignedTransactionSchema };
