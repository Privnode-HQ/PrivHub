/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useState, useRef } from 'react';
import { Button, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
  verifyJSON,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function RequestRateLimit(props) {
  const { t } = useTranslation();

  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    UserGroupUsageLimits: '{}',
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  function onSubmit() {
    const updateArray = compareObjects(inputs, inputsRow);
    if (!updateArray.length) return showWarning(t('你似乎并没有修改什么'));
    const requestQueue = updateArray.map((item) => {
      const value = inputs[item.key];
      return API.put('/api/option/', {
        key: item.key,
        value,
      });
    });
    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (requestQueue.length === 1) {
          if (res.includes(undefined)) return;
        } else if (requestQueue.length > 1) {
          if (res.includes(undefined))
            return showError(t('部分保存失败，请重试'));
        }

        for (let i = 0; i < res.length; i++) {
          if (!res[i].data.success) {
            return showError(res[i].data.message);
          }
        }

        showSuccess(t('保存成功'));
        props.refresh();
      })
      .catch(() => {
        showError(t('保存失败，请重试'));
      })
      .finally(() => {
        setLoading(false);
      });
  }

  useEffect(() => {
    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    refForm.current.setValues(currentInputs);
  }, [props.options]);

  return (
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('用户分组使用限制')}>
            <Row>
              <Col xs={24} sm={16}>
                <Form.TextArea
                  label={t('用户分组使用限制')}
                  placeholder={t(
                    '{\n  "default": {\n    "rpm": 5,\n    "rpm_hide_details": true,\n    "rpd": 20,\n    "tpm": 2000,\n    "tpd": 10000,\n    "hourly": 5000,\n    "daily": 20000,\n    "weekly": 30000,\n    "weekly_hide_details": true,\n    "monthly": 50000\n  },\n  "vip": {\n    "rpm": 20,\n    "rpd": null,\n    "tpm": 20000,\n    "tpd": null,\n    "hourly": null,\n    "daily": null,\n    "weekly": null,\n    "monthly": null\n  }\n}',
                  )}
                  field={'UserGroupUsageLimits'}
                  autosize={{ minRows: 5, maxRows: 15 }}
                  trigger='blur'
                  stopValidateWithError
                  rules={[
                    {
                      validator: (rule, value) => verifyJSON(value),
                      message: t('不是合法的 JSON 字符串'),
                    },
                  ]}
                  extraText={
                    <div>
                      <p>{t('说明：')}</p>
                      <ul>
                        <li>
                          {t(
                            '使用 JSON 对象格式，外层键为用户分组名称，值为包含 rpm、rpd、tpm、tpd、hourly、daily、weekly、monthly 以及可选 *_hide_details 布尔字段的对象。',
                          )}
                        </li>
                        <li>
                          {t(
                            '限制字段只接受整数或 null；null 表示该指标不限制。*_hide_details 只接受 true 或 false。',
                          )}
                        </li>
                        <li>
                          {t(
                            'rpm 和 rpd 表示请求次数限制，tpm 和 tpd 表示 Token 限制，hourly、daily、weekly、monthly 表示预算限制。',
                          )}
                        </li>
                        <li>
                          {t(
                            'hourly、daily、weekly、monthly 使用当前站点额度展示单位进行配置，并会在新的“使用限制”页面展示给用户。',
                          )}
                        </li>
                        <li>
                          {t(
                            '开启 *_hide_details 后，用户侧仅显示消耗百分比，不显示已用、处理中、剩余和总限制。',
                          )}
                        </li>
                        <li>{t('该配置会替代旧的分组速率限制配置。')}</li>
                      </ul>
                    </div>
                  }
                  onChange={(value) => {
                    setInputs({ ...inputs, UserGroupUsageLimits: value });
                  }}
                />
              </Col>
            </Row>
            <Row>
              <Button size='default' onClick={onSubmit}>
                {t('保存使用限制')}
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
