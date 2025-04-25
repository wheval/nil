import { EditorView } from '@codemirror/view';
import {
  CodeField,
  ParagraphSmall,
  TAB_KIND,
  TAG_KIND,
  TAG_SIZE,
  Tab,
  Tabs,
  Tag,
} from '@nilfoundation/ui-kit';
import { Alert } from 'baseui/icon';
import { Key, useState, type FC } from 'react';
import { useStyletron } from 'styletron-react';
import { addressRoute, blockRoute } from '../../routing';
import {
  Divider,
  Info,
  InfoBlock,
  Link,
  addHexPrefix,
  formatShard,
  measure,
} from '../../shared';
import { TokenDisplay } from '../../shared/components/Token';
import type { Transaction } from '../types/Transaction';
import { InlineCopyButton } from './InlineCopyButton';
import InputDataDisplay from './InputDataDisplay';
import { OnChangeHandler, TabsOverrides } from 'baseui/tabs';

type OverviewProps = {
  transaction: Transaction;
};

const styles = {
  infoContainer: {
    display: 'flex',
    flexDirection: 'row',
    alignItems: 'center',
    gap: '1ch',
    height: '1lh',
  },
} as const;

const Bytecode: FC<{ tx: Transaction }> = ({ tx }) => {
  const [css] = useStyletron();
  return tx.method && tx.method.length > 0 ? (
    <CodeField
      extensions={[EditorView.lineWrapping]}
      code={addHexPrefix(tx.method)}
      className={css({ marginTop: '0ch', borderRadius: '16px', padding: '24px' })}
      codeMirrorClassName={css({
        maxHeight: '240px',
        overflowY: 'auto',
        scrollbarWidth: 'thin',
        scrollbarGutter: 'stable',
      })}
    />
  ) : (
    <ParagraphSmall className={css({ margin: '8px'})}>No input data</ParagraphSmall>
  );
};

export const Overview: FC<OverviewProps> = ({ transaction: tx }) => {
  const [css] = useStyletron();
  const [activeKey, setActiveKey] = useState<Key>("0");
    const onChangeHandler: OnChangeHandler = (currentKey) => {
      setActiveKey(currentKey.activeKey);
    };

  return (
    <InfoBlock>
      <Info
        label="Shard + Block height:"
        value={
          <div className={css(styles.infoContainer)}>
            <ParagraphSmall>
              <Link to={blockRoute} params={{ shard: tx.shard_id, id: tx.block_id }}>
                {formatShard(tx.shard_id.toString(), tx.block_id.toString())}
              </Link>
            </ParagraphSmall>
          </div>
        }
      />
      <Info
        label="Hash:"
        value={
          <div className={css(styles.infoContainer)}>
            <ParagraphSmall className={css({ wordBreak: 'break-all', overflowWrap: 'anywhere' })}>
              {addHexPrefix(tx.hash)}
              <InlineCopyButton textToCopy={addHexPrefix(tx.hash)} />
            </ParagraphSmall>
          </div>
        }
      />
      <Info
        label="Nonce:"
        value={
          <div className={css(styles.infoContainer)}>
            <ParagraphSmall>{tx.seqno}</ParagraphSmall>
          </div>
        }
      />
      <Info
        label="Status:"
        value={
          <div className={css(styles.infoContainer)}>
            <ParagraphSmall>
              {tx.success ? (
                'Success'
              ) : (
                <div
                  className={css({
                    display: 'flex',
                    flexDirection: 'row',
                    alignItems: 'center',
                    gap: '0.5ch',
                  })}
                >
                  Failed
                  <Tag kind={TAG_KIND.red} size={TAG_SIZE.s}>
                    <Alert />
                  </Tag>
                </div>
              )}
            </ParagraphSmall>
          </div>
        }
      />
      <Divider />
      <Info
        label="From:"
        value={
          <div className={css(styles.infoContainer)}>
            <ParagraphSmall className={css({ wordBreak: 'break-all', overflowWrap: 'anywhere' })}>
              <Link to={addressRoute} params={{ address: addHexPrefix(tx.from.toLowerCase()) }}>
                {addHexPrefix(tx.from.toLowerCase())}
              </Link>
              <InlineCopyButton textToCopy={addHexPrefix(tx.from.toLowerCase())} />
            </ParagraphSmall>
          </div>
        }
      />
      <Info
        label="To:"
        value={
          <div className={css(styles.infoContainer)}>
            <ParagraphSmall className={css({ wordBreak: 'break-all', overflowWrap: 'anywhere' })}>
              <Link to={addressRoute} params={{ address: addHexPrefix(tx.to).toLowerCase() }}>
                {addHexPrefix(tx.to).toLowerCase()}
              </Link>
              <InlineCopyButton textToCopy={addHexPrefix(tx.to).toLowerCase()} />
            </ParagraphSmall>
          </div>
        }
      />
      <Divider />
      <Info label="Tokens:" value={<TokenDisplay token={tx.token} />} />
      <Divider />
      <Info
        label="Value:"
        value={
          <div className={css(styles.infoContainer)}>
            <ParagraphSmall>{measure(tx.value)}</ParagraphSmall>
          </div>
        }
      />
      <Info
        label="Fee credit:"
        value={
          <div className={css(styles.infoContainer)}>
            <ParagraphSmall>{measure(tx.fee_credit)}</ParagraphSmall>
          </div>
        }
      />
      <Info
        label="Gas used:"
        value={
          <div className={css(styles.infoContainer)}>
            <ParagraphSmall>{`${tx.gas_used ?? 0}`}</ParagraphSmall>
          </div>
        }
      />
      <Divider />
      <Info
        label="Input Data:"
        value={
          <Tabs activeKey={activeKey} onChange={onChangeHandler} overrides={tabsOverrides}>
            <Tab title="Default"  kind={TAB_KIND.secondary}>
              <InputDataDisplay />
            </Tab>
            <Tab title="Bytecode" kind={TAB_KIND.secondary}>
              <Bytecode tx={tx} />
            </Tab>
          </Tabs>
        }
      />
    </InfoBlock>
  );
};

const tabsOverrides: TabsOverrides = {
  TabContent: {
    style: {
      paddingLeft: 0,
      paddingTop: '8px',
      paddingRight: 0,
    },
  },
  TabBar: {
    style: {
      paddingLeft: 0,
      paddingRight: 0,
    },
  },
};
