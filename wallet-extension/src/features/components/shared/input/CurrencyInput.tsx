import { COLORS, Input, Select } from "@nilfoundation/ui-kit";
import { Box, Icon } from "../index.ts";

export interface CurrencyInterface {
  label: string;
  icon: {};
}

interface CurrencyInputProps {
  selectedCurrency: CurrencyInterface;
  currencies: CurrencyInterface[];
  onCurrencyChange: (params: { value: { label: string }[] }) => void;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  error: string;
}

export const CurrencyInput: React.FC<CurrencyInputProps> = ({
  selectedCurrency,
  currencies,
  onCurrencyChange,
  value,
  onChange,
  error,
}) => {
  return (
    <Input
      error={error !== ""}
      placeholder="Enter amount"
      type="number"
      value={value}
      onChange={onChange}
      overrides={{
        Input: {
          style: {
            "::-webkit-outer-spin-button": {
              WebkitAppearance: "none",
              margin: 0,
            },
            "::-webkit-inner-spin-button": {
              WebkitAppearance: "none",
              margin: 0,
            },
            "-moz-appearance": "textfield",
          },
        },
        Root: {
          style: {
            backgroundColor: COLORS.gray700,
            ":hover": {
              backgroundColor: COLORS.gray600,
            },
          },
        },
      }}
      endEnhancer={
        <Box
          $align="center"
          $style={{
            flexDirection: "row",
          }}
        >
          {/* Selected Currency Icon */}
          <Icon
            src={selectedCurrency.icon}
            alt={`${selectedCurrency.label} Icon`}
            size={32}
            iconSize="100%"
            background="transparent"
            margin={"0px 5px"}
          />

          {/* Dropdown */}
          <Select
            options={currencies.map((currency) => ({
              label: currency.label,
              id: currency.label,
            }))}
            searchable={false}
            placeholder=""
            clearable={false}
            overrides={{
              DropdownContainer: {
                style: {
                  width: "100px!important",
                },
              },
              ControlContainer: {
                style: {
                  paddingLeft: "0px!important",
                  paddingRight: "0px!important",
                  backgroundColor: "transparent",
                  boxShadow: "none",
                  ":has(input:focus-within)": {
                    boxShadow: "none",
                  },
                  ":hover": {
                    backgroundColor: "transparent",
                  },
                },
              },
            }}
            onChange={onCurrencyChange}
            value={[{ label: selectedCurrency.label, id: selectedCurrency.label }]}
          />
        </Box>
      }
    />
  );
};
