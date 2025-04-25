import { createEffect, createStore } from 'effector';
import { ethers } from 'ethers';
import { $transaction } from './transaction';
import { $cometaClient } from 'src/features/cometa/model';


// Type for decoded input data
export type DecodedInput = {
  functionName: string;
  methodId: string;
  parameters: { name: string; type: string; value: string }[];
};

// Effect to fetch ABI using Cometa
export const fetchAbiFx = createEffect<string, any, Error>(async (address: string) => {
    const cometa = $cometaClient.getState();
    if (!cometa) throw new Error('Cometa client not initialized');
    if (!address.startsWith('0x')) throw new Error('Invalid hex address format');
    return await cometa.getAbi(address as `0x${string}`);
  });

// Store for the ABI
export const $abi = createStore<any | null>(null)
  .on(fetchAbiFx.doneData, (_, abi) => abi)
  .reset($transaction);

// Store for ABI fetching status
export const $abiStatus = createStore<'idle' | 'loading' | 'success' | 'failed'>('idle')
  .on(fetchAbiFx, () => 'loading')
  .on(fetchAbiFx.done, () => 'success')
  .on(fetchAbiFx.fail, () => 'failed')
  .reset($transaction);

// Store for decoded input data
export const $decodedInput = createStore<DecodedInput | null>(null).on($transaction, (_, tx) => {
  if (!tx || !tx.method || tx.method.length === 0) return null;

  const inputData = addHexPrefix(tx.method);
  if (!ethers.isHexString(inputData) || inputData.length < 10) return null;

  const abi = $abi.getState();
  if (!abi) return null;

  try {
    const iface = new ethers.Interface(abi);
    const selector = inputData.slice(0, 10);
    const func = iface.getFunction(selector);
    if (!func) return null;

    const decoded = iface.decodeFunctionData(func.name, inputData);
    return {
      functionName: func.name,
      methodId: selector,
      parameters: func.inputs.map((input, index) => ({
        name: input.name || `param${index}`,
        type: input.type,
        value: decoded[index].toString(),
      })),
    };
  } catch (error) {
    console.error('Error decoding input:', error);
    return null;
  }
});

$transaction.watch((tx) => {
  if (tx && tx.to) {
    fetchAbiFx(tx.to);
  }
});

// Utility function to add hex prefix
const addHexPrefix = (value: string): string => {
  if (!value) return value;
  return value.startsWith('0x') ? value : `0x${value}`;
};
