import { ParagraphSmall } from '@nilfoundation/ui-kit';
import { useStore } from 'effector-react';
import type { FC } from 'react';
import type { Transaction } from '../types/Transaction';
import { $decodedInput, $abiStatus, type DecodedInput } from '../models/transactionInput';
import { $transaction } from '../models/transaction';
import { useStyletron } from 'styletron-react';

const InputDataDisplay: FC = () => {
  const [css] = useStyletron();
  const tx = useStore($transaction) as Transaction | null;
  const decodedInput = useStore($decodedInput) as DecodedInput | null;
  const abiStatus = useStore($abiStatus);

  if (!tx || !tx.method || tx.method.length === 0) {
    return <ParagraphSmall>No input data</ParagraphSmall>;
  }

  if (abiStatus === 'loading') {
    return <ParagraphSmall>Decoding input data...</ParagraphSmall>;
  }

  if (abiStatus === 'failed') {
    return <ParagraphSmall>Failed to fetch ABI, unable to decode</ParagraphSmall>;
  }

  if (decodedInput) {
    return (
      <div>
        <ParagraphSmall>
          <span className={css({color: "#8A8A8A"})}>Function:</span> {decodedInput.functionName}
        </ParagraphSmall>
        <ParagraphSmall>
          <span className={css({color: "#8A8A8A"})}>Method ID:</span> {decodedInput.methodId}
        </ParagraphSmall>
        <ParagraphSmall>
          <span className={css({color: "#8A8A8A"})}>Parameters:</span>
        </ParagraphSmall>
        {decodedInput.parameters.map((param: { name: string; type: string; value: string }, index: number) => (
          <ParagraphSmall key={`${index}-${param.name}`}>
            {param.name} ({param.type}): {param.value}
          </ParagraphSmall>
        ))}
      </div>
    );
  }

  return <ParagraphSmall>Unable to decode input data</ParagraphSmall>;
};

export default InputDataDisplay;
